#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
KIND_CLUSTER_NAME="${KIND_CLUSTER_NAME:-podbuster-chaos}"
PODBUSTER_NAMESPACE="${PODBUSTER_NAMESPACE:-podbuster-chaos}"
KIND_CONTEXT="kind-${KIND_CLUSTER_NAME}"
TARGETS_MANIFEST="${ROOT_DIR}/deploy/kind-chaos-playground/targets.yaml"

require_command() {
	local command_name="$1"
	local install_hint="$2"

	if ! command -v "${command_name}" >/dev/null 2>&1; then
		printf 'error: %s is required. %s\n' "${command_name}" "${install_hint}" >&2
		exit 1
	fi
}

runtime_ready() {
	local provider="$1"

	case "${provider}" in
		docker)
			command -v docker >/dev/null 2>&1 && docker info >/dev/null 2>&1
			;;
		podman)
			command -v podman >/dev/null 2>&1 && podman info >/dev/null 2>&1
			;;
		*)
			return 1
			;;
	esac
}

select_kind_provider() {
	require_command "kind" "Install kind from https://kind.sigs.k8s.io/docs/user/quick-start/."
	require_command "kubectl" "Install kubectl from https://kubernetes.io/docs/tasks/tools/."

	if [[ -n "${KIND_EXPERIMENTAL_PROVIDER:-}" ]]; then
		case "${KIND_EXPERIMENTAL_PROVIDER}" in
			docker|podman)
				if ! runtime_ready "${KIND_EXPERIMENTAL_PROVIDER}"; then
					printf 'error: KIND_EXPERIMENTAL_PROVIDER=%s is set, but that runtime is not available or not running.\n' "${KIND_EXPERIMENTAL_PROVIDER}" >&2
					exit 1
				fi
				printf 'Using kind provider: %s\n' "${KIND_EXPERIMENTAL_PROVIDER}"
				return
				;;
			*)
				printf 'error: unsupported KIND_EXPERIMENTAL_PROVIDER=%s. Use docker or podman.\n' "${KIND_EXPERIMENTAL_PROVIDER}" >&2
				exit 1
				;;
		esac
	fi

	if runtime_ready docker; then
		unset KIND_EXPERIMENTAL_PROVIDER
		printf 'Using kind provider: docker\n'
		return
	fi

	if runtime_ready podman; then
		export KIND_EXPERIMENTAL_PROVIDER=podman
		printf 'Using kind provider: podman\n'
		return
	fi

	printf 'error: kind needs a running container runtime. Start Docker, or start Podman and run with KIND_EXPERIMENTAL_PROVIDER=podman.\n' >&2
	exit 1
}

cluster_exists() {
	kind get clusters 2>/dev/null | grep -Fxq "${KIND_CLUSTER_NAME}"
}

ensure_cluster() {
	if cluster_exists; then
		printf 'Reusing kind cluster: %s\n' "${KIND_CLUSTER_NAME}"
	else
		printf 'Creating kind cluster: %s\n' "${KIND_CLUSTER_NAME}"
		kind create cluster --name "${KIND_CLUSTER_NAME}"
	fi
}

ensure_namespace() {
	kubectl --context "${KIND_CONTEXT}" create namespace "${PODBUSTER_NAMESPACE}" --dry-run=client -o yaml \
		| kubectl --context "${KIND_CONTEXT}" apply -f -
	kubectl --context "${KIND_CONTEXT}" label namespace "${PODBUSTER_NAMESPACE}" \
		app.kubernetes.io/part-of=podbuster \
		podbuster.github.com/chaos-playground=true \
		--overwrite >/dev/null
}

apply_targets() {
	kubectl --context "${KIND_CONTEXT}" -n "${PODBUSTER_NAMESPACE}" apply -f "${TARGETS_MANIFEST}"
}

wait_for_targets() {
	kubectl --context "${KIND_CONTEXT}" -n "${PODBUSTER_NAMESPACE}" wait \
		--for=condition=Available \
		deployment \
		-l app.kubernetes.io/part-of=podbuster,podbuster.github.com/chaos-playground=true \
		--timeout=120s
}

delete_namespace_if_present() {
	if kubectl --context "${KIND_CONTEXT}" get namespace "${PODBUSTER_NAMESPACE}" >/dev/null 2>&1; then
		kubectl --context "${KIND_CONTEXT}" delete namespace "${PODBUSTER_NAMESPACE}" --wait=true
	fi
}
