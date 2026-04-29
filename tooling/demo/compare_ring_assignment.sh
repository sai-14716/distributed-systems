#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$ROOT_DIR"
export PYTHONPATH="$ROOT_DIR"

# Machine C local containers
LOCAL_NODES="node4 node5"

LB_URL="${1:-$(python3 tooling/cluster_helper.py get_url load_balancers node4)}"
NODE_TO_TOGGLE="${2:-node5}"
OUT_DIR="${3:-/tmp}"
WAIT_SECS="${WAIT_SECS:-5}"

# Verify the node to toggle is local
is_local=false
for ln in $LOCAL_NODES; do
  if [[ "$NODE_TO_TOGGLE" == "$ln" ]]; then
    is_local=true
    break
  fi
done

if [[ "$is_local" != "true" ]]; then
  echo "Error: $NODE_TO_TOGGLE is not a local container on this machine." >&2
  echo "This machine only controls: $LOCAL_NODES" >&2
  exit 1
fi

mkdir -p "$OUT_DIR"

before_file="$OUT_DIR/owners.before.tsv"
after_stop_file="$OUT_DIR/owners.after-stop.tsv"
after_recover_file="$OUT_DIR/owners.after-recover.tsv"

########################################
# Extract backend -> owner
########################################
extract_raw() {
  local out_file="$1"
  curl -sf "$LB_URL/admin/status" | jq -r '
    .load_view.backends
    | to_entries
    | sort_by(.key)
    | .[]
    | [
        .key,
        (.value.view.owner_id // "-")
      ]
    | @tsv
  ' > "$out_file"
}

########################################
# Print node -> backends (with highlight)
########################################
print_grouped() {
  local raw_file="$1"
  local changed_file="${2:-}"

  awk -v changed_file="$changed_file" '
    BEGIN {
      FS="\t"

      # load changed backends if provided
      if (changed_file != "") {
        while ((getline line < changed_file) > 0) {
          changed[line] = 1
        }
      }
    }
    {
      backend=$1
      node=$2
      nodes[node] = nodes[node] " " backend
    }
    END {
      for (n in nodes) {
        printf "%-10s ", n
        split(nodes[n], arr, " ")
        for (i in arr) {
          b = arr[i]
          if (b == "") continue
          if (b in changed) {
            printf "[%s]* ", b
          } else {
            printf "%s ", b
          }
        }
        printf "\n"
      }
    }
  ' "$raw_file" | sort
}

########################################
# Compute reassigned backends
########################################
compute_changed() {
  local before="$1"
  local after="$2"
  local out="$3"

  join -t $'\t' -a1 -a2 -e "-" -o '0,1.2,2.2' \
    <(sort "$before") \
    <(sort "$after") \
  | awk -F'\t' '$2 != $3 {print $1}' > "$out"
}

########################################
# BASELINE
########################################
echo "[ring-demo] baseline from ${LB_URL}/admin/status"

extract_raw "$before_file"

echo "[ring-demo] node -> backends"
print_grouped "$before_file"

########################################
# STOP NODE
########################################
echo
echo "[ring-demo] stopping $NODE_TO_TOGGLE"
docker compose stop "$NODE_TO_TOGGLE" >/dev/null
sleep "$WAIT_SECS"

extract_raw "$after_stop_file"

changed_stop="$OUT_DIR/changed.stop"
compute_changed "$before_file" "$after_stop_file" "$changed_stop"

echo
echo "[ring-demo] after stop (highlight reassigned)"
print_grouped "$after_stop_file" "$changed_stop"

########################################
# RECOVER NODE
########################################
echo
echo "[ring-demo] recovering $NODE_TO_TOGGLE"
docker compose up -d "$NODE_TO_TOGGLE" >/dev/null
sleep "$WAIT_SECS"

extract_raw "$after_recover_file"

changed_recover="$OUT_DIR/changed.recover"
compute_changed "$after_stop_file" "$after_recover_file" "$changed_recover"

echo
echo "[ring-demo] after recover"
print_grouped "$after_recover_file" "$changed_recover"

########################################
# OUTPUT FILES
########################################
echo
echo "[ring-demo] wrote:"
echo "  $before_file"
echo "  $after_stop_file"
echo "  $after_recover_file"