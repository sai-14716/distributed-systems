#!/usr/bin/env bash

cluster_config_file() {
  printf '%s\n' "${CLUSTER_CONFIG:-$ROOT_DIR/cluster_config.yaml}"
}

cluster_py() {
  python3 - "$@" <<'PY'
import json
import sys

cfg_path = sys.argv[1]
mode = sys.argv[2]
args = sys.argv[3:]

try:
    with open(cfg_path, encoding="utf-8") as f:
        cfg = json.load(f)
except FileNotFoundError:
    raise SystemExit(f"cluster config not found: {cfg_path}")
except Exception as exc:
    raise SystemExit(f"failed to read cluster config {cfg_path}: {exc}")

laptops = cfg.get("laptops") or {}

def resolve(service_type, name):
    info = cfg[service_type][name]
    ip = laptops[info["laptop"]]
    return ip, info

def names(service_type):
    return list((cfg.get(service_type) or {}).keys())

def url(service_type, name, endpoint=""):
    ip, info = resolve(service_type, name)
    base = f"http://{ip}:{info['port']}"
    endpoint = endpoint.lstrip("/")
    return f"{base}/{endpoint}" if endpoint else base

if mode == "names":
    print(" ".join(names(args[0])))
elif mode == "csv_urls":
    service_type = args[0]
    endpoint = args[1] if len(args) > 1 else ""
    print(",".join(url(service_type, name, endpoint) for name in names(service_type)))
elif mode == "url":
    endpoint = args[2] if len(args) > 2 else ""
    print(url(args[0], args[1], endpoint))
elif mode == "first_url":
    service_type = args[0]
    endpoint = args[1] if len(args) > 1 else ""
    svc_names = names(service_type)
    if not svc_names:
        raise SystemExit(f"no services configured in {service_type}")
    print(url(service_type, svc_names[0], endpoint))
elif mode == "discovery_http":
    disc = cfg["discovery"]
    ip = laptops[disc["laptop"]]
    endpoint = args[0].lstrip("/") if args else ""
    base = f"http://{ip}:{disc['http_port']}"
    print(f"{base}/{endpoint}" if endpoint else base)
elif mode == "discovery_udp":
    disc = cfg["discovery"]
    ip = laptops[disc["laptop"]]
    print(f"{ip}:{disc['udp_port']}")
else:
    raise SystemExit(f"unknown cluster config mode: {mode}")
PY
}

cluster_names() {
  cluster_py "$(cluster_config_file)" names "$1"
}

cluster_url() {
  local service_type="$1"
  local name="$2"
  local endpoint="${3:-}"
  cluster_py "$(cluster_config_file)" url "$service_type" "$name" "$endpoint"
}

cluster_first_url() {
  local service_type="$1"
  local endpoint="${2:-}"
  cluster_py "$(cluster_config_file)" first_url "$service_type" "$endpoint"
}

cluster_csv_urls() {
  local service_type="$1"
  local endpoint="${2:-}"
  cluster_py "$(cluster_config_file)" csv_urls "$service_type" "$endpoint"
}

cluster_discovery_http_url() {
  local endpoint="${1:-}"
  cluster_py "$(cluster_config_file)" discovery_http "$endpoint"
}

cluster_discovery_udp_addr() {
  cluster_py "$(cluster_config_file)" discovery_udp
}
