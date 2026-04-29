#!/usr/bin/env bash
set -euo pipefail

# Validate live LB controller probe assignments against pkg/lb/controller_probe_config.json.
#
# Usage:
#   bash tooling/demo/check_probe_overlap.sh [wait_seconds]
#
# Optional env:
#   CLUSTER_CONFIG=cluster_config.yaml
#   PROBE_CONFIG=pkg/lb/controller_probe_config.json
#   LB_URLS=node1=http://10.0.0.1:8001,node2=http://10.0.0.2:8002

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$ROOT_DIR"

WAIT_SECONDS="${1:-20}"
PROBE_CONFIG="${PROBE_CONFIG:-pkg/lb/controller_probe_config.json}"
CLUSTER_CONFIG="${CLUSTER_CONFIG:-cluster_config.yaml}"

python3 - <<'PY' "$WAIT_SECONDS" "$CLUSTER_CONFIG" "$PROBE_CONFIG" "${LB_URLS:-}"
import json
import sys
import time
import urllib.request
from concurrent.futures import ThreadPoolExecutor, as_completed

wait_seconds = int(sys.argv[1])
cluster_config_path = sys.argv[2]
probe_config_path = sys.argv[3]
lb_urls_env = sys.argv[4].strip()

def load_json(path):
    with open(path, encoding="utf-8") as f:
        return json.load(f)

def normalize_backend_ids(assignments):
    out = {}
    for node, backends in assignments.items():
        vals = []
        for backend in backends:
            backend = str(backend).strip()
            if not backend:
                continue
            if backend.isdigit():
                backend = f"backend-{backend}"
            vals.append(backend)
        out[str(node)] = vals
    return out

probe_cfg = load_json(probe_config_path)
assignments = normalize_backend_ids(probe_cfg.get("assignments") or {})

backend_to_probers = {}
for node, backends in assignments.items():
    for backend in backends:
        backend_to_probers.setdefault(backend, set()).add(node)

overlapped = {b: sorted(nodes) for b, nodes in backend_to_probers.items() if len(nodes) > 1}
if not overlapped:
    raise SystemExit("FAIL: probe config has no overlapping backend assignments")

def lb_urls_from_env(raw):
    out = {}
    for entry in raw.split(","):
        entry = entry.strip()
        if not entry:
            continue
        if "=" not in entry:
            raise SystemExit(f"invalid LB_URLS entry {entry!r}; expected node=url")
        node, url = entry.split("=", 1)
        out[node.strip()] = url.strip().rstrip("/")
    return out

def lb_urls_from_cluster_config(path):
    cfg = load_json(path)
    laptops = cfg.get("laptops") or {}
    out = {}
    for node, info in (cfg.get("load_balancers") or {}).items():
        laptop = info.get("laptop")
        ip = laptops.get(laptop)
        port = info.get("port")
        if not ip or not port:
            continue
        out[node] = f"http://{ip}:{port}"
    return out

lb_urls = lb_urls_from_env(lb_urls_env) if lb_urls_env else lb_urls_from_cluster_config(cluster_config_path)
if not lb_urls:
    raise SystemExit("FAIL: no LB URLs found. Set LB_URLS or check cluster_config.yaml")

live_nodes = set(lb_urls)
expected_for_live_nodes = {node: backends for node, backends in assignments.items() if node in live_nodes}
if not expected_for_live_nodes:
    raise SystemExit(f"FAIL: none of the probe-config nodes are in live LB URLs: {sorted(live_nodes)}")

def fetch_snapshot(node, base_url):
    url = base_url.rstrip("/") + "/admin/load-view"
    with urllib.request.urlopen(url, timeout=2.0) as resp:
        return node, json.loads(resp.read().decode("utf-8"))

def fetch_all():
    out = {}
    errors = {}
    with ThreadPoolExecutor(max_workers=max(1, len(lb_urls))) as pool:
        futures = [pool.submit(fetch_snapshot, node, url) for node, url in lb_urls.items()]
        for fut in as_completed(futures):
            try:
                node, snap = fut.result()
                out[node] = snap
            except Exception as exc:
                errors[str(exc)] = errors.get(str(exc), 0) + 1
    return out, errors

def entry_for(snap, backend):
    return (snap.get("backends") or {}).get(backend) or {}

def is_local_prober(snap, backend):
    return entry_for(snap, backend).get("local_role") == "prober"

def has_report(snap, backend, prober):
    report = ((entry_for(snap, backend).get("reports") or {}).get(prober) or {})
    status = str(report.get("status") or "").lower()
    return status in {"up", "down"}

deadline = time.time() + wait_seconds
last_snaps = {}
last_errors = {}

while time.time() <= deadline:
    snaps, errors = fetch_all()
    last_snaps = snaps
    last_errors = errors

    if set(snaps) >= set(expected_for_live_nodes):
        local_ok = True
        for node, backends in expected_for_live_nodes.items():
            snap = snaps.get(node) or {}
            for backend in backends:
                if not is_local_prober(snap, backend):
                    local_ok = False
                    break
            if not local_ok:
                break

        report_ok = True
        for backend, probers in backend_to_probers.items():
            live_probers = sorted(probers & live_nodes)
            if not live_probers:
                continue
            for observer, snap in snaps.items():
                for prober in live_probers:
                    if not has_report(snap, backend, prober):
                        report_ok = False
                        break
                if not report_ok:
                    break
            if not report_ok:
                break

        if local_ok and report_ok:
            break

    time.sleep(1)

print("Probe config overlaps:")
for backend, nodes in sorted(overlapped.items()):
    print(f"  {backend}: {', '.join(nodes)}")

print("\nLB endpoints checked:")
for node, url in sorted(lb_urls.items()):
    status = "ok" if node in last_snaps else "missing"
    print(f"  {node}: {url} [{status}]")

failures = []
for node, backends in sorted(expected_for_live_nodes.items()):
    snap = last_snaps.get(node) or {}
    for backend in backends:
        if not is_local_prober(snap, backend):
            failures.append(f"{node} is not locally probing expected {backend}")

for backend, probers in sorted(backend_to_probers.items()):
    live_probers = sorted(probers & live_nodes)
    if not live_probers:
        continue
    for observer, snap in sorted(last_snaps.items()):
        for prober in live_probers:
            if not has_report(snap, backend, prober):
                failures.append(f"{observer} has no gossiped report for {backend} from {prober}")

if last_errors:
    print("\nFetch errors seen:")
    for err, count in sorted(last_errors.items()):
        print(f"  {count}x {err}")

if failures:
    print("\nFAIL:")
    for msg in failures[:40]:
        print(f"  - {msg}")
    if len(failures) > 40:
        print(f"  ... {len(failures) - 40} more")
    raise SystemExit(1)

print("\nPASS: live LBs match overlapping probe assignments and gossip reports are visible.")
PY
