#!/usr/bin/env bash
#
# run locally
#

set -o errexit
set -o pipefail
set -o nounset

if [ -f ".env" ]; then
	export $(cat .env)
fi

~/go/bin/air
