#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/kind-lib.sh"

select_kind_provider
ensure_cluster
delete_namespace_if_present
ensure_namespace
apply_targets
wait_for_targets

printf 'Chaos playground reset: context=%s namespace=%s\n' "${KIND_CONTEXT}" "${PODBUSTER_NAMESPACE}"
