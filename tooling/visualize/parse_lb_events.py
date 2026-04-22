#!/usr/bin/env python3
"""
Parse Go LB logs from stdin and emit JSON Lines for visualization.

Typical usage:
  docker compose logs -f node1 node2 node3 node4 node5 | python3 tooling/visualize/parse_lb_events.py > /tmp/lb_events.jsonl

It extracts lines that contain:
  [lb-load] event=...
  [lb-gossip] event=...
and converts key=value pairs into a JSON object.
"""

import json
import re
import sys


KV_RE = re.compile(r"([a-zA-Z0-9_]+)=([^ ]+)")


def parse_line(line: str):
    line = line.strip()
    if "[lb-load]" not in line and "[lb-gossip]" not in line:
        return None

    kind = "lb-load" if "[lb-load]" in line else "lb-gossip"

    # docker compose logs prefix looks like: "node1  | 2026/04/19 12:34:56 ..."
    # Go log prefix looks like: "2026/04/19 12:34:56 ..."
    ts = None
    m = re.search(r"(\d{4}/\d{2}/\d{2} \d{2}:\d{2}:\d{2})", line)
    if m:
        ts = m.group(1)

    obj = {"kind": kind}
    if ts:
        obj["ts"] = ts

    for k, v in KV_RE.findall(line):
        # normalize ints where obvious
        if v.isdigit():
            obj[k] = int(v)
        else:
            obj[k] = v
    return obj


def main():
    for raw in sys.stdin:
        obj = parse_line(raw)
        if obj is None:
            continue
        sys.stdout.write(json.dumps(obj) + "\n")
        sys.stdout.flush()


if __name__ == "__main__":
    main()

