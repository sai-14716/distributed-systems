#!/bin/sh
set -eu

DNS_ADDR="${DISCOVERY_DNS_ADDR:-discovery:6699}"
LISTEN_ADDR="${DISCOVERY_CLIENT_LISTEN:-127.0.0.1:6700}"

python3 /app/service_discovery_client.py \
  --listen "${LISTEN_ADDR}" \
  --dns "${DNS_ADDR}" \
  --cache-ttl-ms "${DISCOVERY_CACHE_TTL_MS:-500}" \
  --max-tries "${DISCOVERY_MAX_TRIES:-3}" &

K6_SCRIPT="${K6_SCRIPT:-/app/k6/stress-test.js}"
if [ "${DISABLE_PROXY:-0}" = "1" ]; then
    unset HTTP_PROXY
    unset HTTPS_PROXY
    export NO_PROXY="*"
else
    export HTTP_PROXY="http://${LISTEN_ADDR}"
    export NO_PROXY=""
fi

if [ "${DISABLE_PROXY:-0}" != "1" ]; then
    # Warm up: wait for proxy + discovery + at least one node to be reachable.
    python3 - <<'PY'
import os, time, urllib.request

proxy = os.environ["HTTP_PROXY"]
target = "http://api.service.com:8000/chat"
handler = urllib.request.ProxyHandler({"http": proxy, "https": proxy})
opener = urllib.request.build_opener(handler)

deadline = time.time() + 15
last_err = None
while time.time() < deadline:
        try:
                with opener.open(target, timeout=1.0) as r:
                        if r.status == 200:
                                break
        except Exception as e:
                last_err = e
                time.sleep(0.1)
else:
        print(f"[entrypoint] warmup failed: {last_err}")
PY
fi

exec k6 run "${K6_SCRIPT}"
