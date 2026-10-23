#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/kind-lib.sh"

select_kind_provider

if cluster_exists; then
	printf 'Deleting kind cluster: %s\n' "${KIND_CLUSTER_NAME}"
	kind delete cluster --name "${KIND_CLUSTER_NAME}"
else
	printf 'Kind cluster does not exist: %s\n' "${KIND_CLUSTER_NAME}"
fi
