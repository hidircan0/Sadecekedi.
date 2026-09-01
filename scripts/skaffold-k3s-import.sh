#!/usr/bin/env bash
# Skaffold build hook: Docker'daki image'ları deploy tag'iyle k3s'e aktarır.
# Skaffold local dev'de manifest'teki :latest yerine image digest tag kullanır.
set -euo pipefail

import_one() {
  local ref="$1"
  if ! docker image inspect "$ref" >/dev/null 2>&1; then
    echo "Atlanıyor (Docker'da yok): $ref"
    return 0
  fi

  local digest tag repo
  digest=$(docker inspect --format='{{.Id}}' "$ref")
  tag="${digest#sha256:}"
  repo="${ref%%:*}"

  docker tag "$ref" "${repo}:${tag}"
  echo "→ k3s import: ${repo}:${tag}"
  docker save "${repo}:${tag}" | sudo k3s ctr images import -
}

IMAGES=(
  "docker.io/library/sadecekedi-backend:latest"
  "docker.io/library/sadecekedi-yolo:latest"
)

for ref in "${IMAGES[@]}"; do
  import_one "$ref"
done

echo "✓ k3s image import tamam"
