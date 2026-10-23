KIND_CLUSTER_NAME ?= podbuster-chaos
PODBUSTER_NAMESPACE ?= podbuster-chaos

export KIND_CLUSTER_NAME
export PODBUSTER_NAMESPACE

.PHONY: kind-up kind-reset kind-down run-kind-chaos run-mock check lint fmt test

GOLANGCI_LINT_VERSION := v2.14.0
GOLANGCI_LINT := $(CURDIR)/.bin/golangci-lint-$(GOLANGCI_LINT_VERSION)

check: lint test

lint: $(GOLANGCI_LINT)
	go mod tidy -diff
	$(GOLANGCI_LINT) fmt --diff
	$(GOLANGCI_LINT) run ./...

fmt: $(GOLANGCI_LINT)
	go mod tidy
	$(GOLANGCI_LINT) fmt
	$(GOLANGCI_LINT) run --fix ./...

test:
	go build ./...
	go test -race ./...

$(GOLANGCI_LINT):
	GOBIN=$(CURDIR)/.bin go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)
	mv $(CURDIR)/.bin/golangci-lint $@

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
