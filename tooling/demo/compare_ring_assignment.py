#!/usr/bin/env python3
from __future__ import annotations

import argparse
import json
import zlib


def parse_csv(value: str) -> list[str]:
    items = [v.strip() for v in value.split(",") if v.strip()]
    # keep output clean and deterministic
    seen: set[str] = set()
    formatted: list[str] = []
    for item in items:
        if item not in seen:
            seen.add(item)
            formatted.append(item)
    return formatted


def build_ring(peers: list[str], replicas: int) -> tuple[list[tuple[int, str]], list[str]]:
    if replicas < 1:
        replicas = 1
    ids = sorted(set(peers))
    if not ids:
        raise ValueError("at least one peer is required")
    points: list[tuple[int, str]] = []
    for peer_id in ids:
        for i in range(replicas):
            key = f"{peer_id}#{i}".encode("utf-8")
            points.append((zlib.crc32(key) & 0xFFFFFFFF, peer_id))
    points.sort(key=lambda p: (p[0], p[1]))
    return points, ids


def owner(points: list[tuple[int, str]], node_ids: list[str], backend_id: str) -> str:
    if len(node_ids) == 1:
        return node_ids[0]

    h = zlib.crc32(backend_id.encode("utf-8")) & 0xFFFFFFFF

    lo = 0
    hi = len(points)
    while lo < hi:
        mid = (lo + hi) // 2
        if points[mid][0] >= h:
            hi = mid
        else:
            lo = mid + 1
    idx = lo if lo < len(points) else 0

    return points[idx][1]


def main() -> int:
    parser = argparse.ArgumentParser(description="Compare ring ownership before/after LB membership changes")
    parser.add_argument("--before-peers", required=True, help="comma-separated peer IDs before change")
    parser.add_argument("--after-peers", required=True, help="comma-separated peer IDs after change")
    parser.add_argument("--backends", default="", help="comma-separated backend IDs")
    parser.add_argument("--config", default="cluster_config.yaml", help="cluster topology config")
    parser.add_argument("--replicas", type=int, default=50, help="virtual nodes per peer (default: 50)")
    args = parser.parse_args()

    before_peers = parse_csv(args.before_peers)
    after_peers = parse_csv(args.after_peers)
    if args.backends:
        backends = parse_csv(args.backends)
    else:
        try:
            with open(args.config, encoding="utf-8") as f:
                cfg = json.load(f)
            backends = list((cfg.get("backends") or {}).keys())
        except Exception as exc:
            raise SystemExit(f"failed to read backend list from {args.config}: {exc}")

    if not before_peers or not after_peers or not backends:
        raise SystemExit("before-peers, after-peers, and backends must be non-empty")

    before_ring, before_ids = build_ring(before_peers, args.replicas)
    after_ring, after_ids = build_ring(after_peers, args.replicas)

    print("backend\tbefore_owner\tafter_owner\towner_changed")
    owner_changed = 0
    for backend_id in backends:
        b = owner(before_ring, before_ids, backend_id)
        a = owner(after_ring, after_ids, backend_id)
        changed = b != a
        if changed:
            owner_changed += 1
        print(f"{backend_id}\t{b}\t{a}\t{str(changed).lower()}")

    total = len(backends)
    print()
    print(f"summary: backends={total} owner_changed={owner_changed}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
