#!/usr/bin/env bash
#
# run locally but in docker

set -o errexit
set -o pipefail
set -o nounset

docker build -t regexplanet-go .
docker run \
	--publish 4000:4000 \
	--expose 4000 \
	--env PORT='4000' \
	--env COMMIT=$(git rev-parse --short HEAD) \
	--env LASTMOD=$(date -u +%Y-%m-%dT%H:%M:%SZ) \
	regexplanet-go
