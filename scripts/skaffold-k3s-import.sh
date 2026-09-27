#!/usr/bin/env bash
# Docker image'ını k3s'e, pod'ların istediği tag ile aktarır (:latest).
set -euo pipefail

find_k3s() {
  if command -v k3s >/dev/null 2>&1; then
    command -v k3s
    return
  fi
  local p
  for p in /usr/local/bin/k3s /usr/bin/k3s; do
    if [ -x "$p" ]; then
      echo "$p"
      return
    fi
  done
  return 1
}

if ! K3S="$(find_k3s)"; then
  echo "k3s bu makinede yok." >&2
  echo "Cluster kur: curl -sfL https://get.k3s.io | sh -" >&2
  exit 1
fi

k3s_ctr() {
  sudo "$K3S" ctr "$@"
}

in_k3s() {
  local ref="$1"
  k3s_ctr -n k8s.io images ls -q 2>/dev/null | grep -Fq "$ref"
}

import_one() {
  local img="$1"
  if ! docker image inspect "$img" >/dev/null 2>&1; then
    echo "Docker'da yok, pull: $img"
    docker pull "$img" || {
      echo "pull başarısız: $img" >&2
      return 0
    }
  fi

  # App images are rebuilt often; :latest already in k3s must be overwritten.
  if [[ "$img" != *sadecekedi* ]] && in_k3s "$img"; then
    echo "→ k3s'te var: $img"
    return 0
  fi

  echo "→ k3s import: $img"
  docker save "$img" | k3s_ctr images import -
}

IMAGES=(
  "docker.io/library/sadecekedi-backend:latest"
  "docker.io/library/sadecekedi-yolo:latest"
  "adobe/s3mock:latest"
  "busybox:latest"
  "postgres:15-alpine"
)

docker tag sadecekedi-backend:latest docker.io/library/sadecekedi-backend:latest 2>/dev/null || true
docker tag sadecekedi-yolo:latest docker.io/library/sadecekedi-yolo:latest 2>/dev/null || true

for img in "${IMAGES[@]}"; do
  import_one "$img"
done

echo "✓ k3s image import tamam"
