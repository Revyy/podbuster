# Kind chaos playground

The chaos playground is a disposable local kind cluster with one namespace full of safe Pods.
Use it to try chaos mode without pointing Podbuster at production infrastructure.

## Prerequisites

- `kind`
- `kubectl`
- Docker, or Podman with kind's Podman provider

The scripts do not install these tools. They fail with a clear error if `kind`, `kubectl`, or a
running container runtime is missing.

## Create the playground

```bash
make kind-up
```

This creates or reuses a kind cluster named `podbuster-chaos`, creates a namespace named
`podbuster-chaos`, applies the safe target workloads, and waits until the Deployments are available.

The targets give every enemy archetype and tier something to shoot. The median request is 50m
CPU, so tiers come out as listed:

| Workload | Pods | Requests (CPU/memory) | Archetype / Tier |
|---|---|---|---|
| Deployment `target-1` | 1 | none | Probe / Light |
| Deployment `target-2` | 1 | 10m/16Mi | Probe / Light |
| Deployments `target-3`, `target-4` | 1 each | 50m/64Mi | Probe / Standard |
| Deployment `target-5` | 1 | 250m/256Mi | Probe / Heavy |
| StatefulSet `target-sts` | 2 | 50m/64Mi | Turret / Standard |
| DaemonSet `target-ds` | 1 on single-node kind | 10m/16Mi | Sentinel / Light |
| Job `target-job` | 2 | 250m/256Mi | Dart / Heavy |

The Job runs `pause`, so its Pods never complete on their own, and `backoffLimit: 100` keeps chaos
deletions from exhausting its retries.

For Podman, set kind's provider explicitly:

```bash
KIND_EXPERIMENTAL_PROVIDER=podman make kind-up
```

You can override the cluster and namespace names:

```bash
KIND_CLUSTER_NAME=my-podbuster PODBUSTER_NAMESPACE=my-podbuster make kind-up
```

## Run Podbuster against the playground

```bash
make run-kind-chaos
```

The target expands to:

```bash
go run . --context kind-podbuster-chaos --namespace podbuster-chaos
```

Podbuster still starts at the title screen. Select chaos mode in-game and confirm the namespace
before any Pod deletion can happen.

To watch Kubernetes replace Pods while you play, run this in another terminal:

```bash
kubectl --context kind-podbuster-chaos -n podbuster-chaos get pods -w
```

## Reset after playing

```bash
make kind-reset
```

This deletes and recreates only the playground namespace, then reapplies the target workloads.
It preserves the kind cluster.

## Delete the playground

```bash
make kind-down
```

This deletes only the kind cluster named by `KIND_CLUSTER_NAME`, which defaults to
`podbuster-chaos`.

## Safety notes

- The launch command passes `--context kind-podbuster-chaos` explicitly, so Podbuster does not rely
  on your current `kubectl` context.
- The default namespace is dedicated to the playground and labelled
  `podbuster.github.com/chaos-playground=true`.
- Do not reuse production namespaces or production cluster names for this setup.
- If you use Podman, use the same `KIND_EXPERIMENTAL_PROVIDER=podman` setting for `kind-up`,
  `kind-reset`, `run-kind-chaos`, and `kind-down`.
