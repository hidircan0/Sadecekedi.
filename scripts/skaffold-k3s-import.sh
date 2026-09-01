#!/usr/bin/env bash
# Docker image'ını k3s'e aktarır; aynı digest zaten varsa tar import atlanır.
set -euo pipefail

already_in_k3s() {
  local ref="$1"
  sudo k3s ctr -n k8s.io images ls -q 2>/dev/null | grep -Fxq "$ref"
}

import_one() {
  local img="$1"
  if ! docker image inspect "$img" >/dev/null 2>&1; then
    echo "Atlanıyor (Docker'da yok): $img"
    return 0
  fi

  local digest tag repo
  digest=$(docker inspect --format='{{.Id}}' "$img")
  tag="${digest#sha256:}"
  repo="${img%%:*}"
  local digest_ref="${repo}:${tag}"

  docker tag "$img" "$digest_ref"

  if already_in_k3s "$digest_ref"; then
    echo "→ k3s'te var, import yok: $digest_ref"
    return 0
  fi

  echo "→ k3s import: $digest_ref"
  docker save "$digest_ref" | sudo k3s ctr images import -
}

IMAGES=(
  "docker.io/library/sadecekedi-backend:latest"
  "docker.io/library/sadecekedi-yolo:latest"
)

for img in "${IMAGES[@]}"; do
  import_one "$img"
done

echo "✓ k3s image import tamam"
