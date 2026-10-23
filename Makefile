KIND_CLUSTER_NAME ?= podbuster-chaos
PODBUSTER_NAMESPACE ?= podbuster-chaos

export KIND_CLUSTER_NAME
export PODBUSTER_NAMESPACE

.PHONY: kind-up kind-reset kind-down run-kind-chaos run-mock

kind-up:
	./scripts/kind-up.sh

kind-reset:
	./scripts/kind-reset.sh

kind-down:
	./scripts/kind-down.sh

run-kind-chaos: kind-up
	go run . --context kind-$(KIND_CLUSTER_NAME) --namespace $(PODBUSTER_NAMESPACE)

run-mock:
	go run . --mock
