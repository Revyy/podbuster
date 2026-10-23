# Choosing mock and playground variety

Type: grilling
Status: resolved
Blocked by: 09

## Question

What mix of Pods should mock mode and the chaos playground provide, so every archetype and tier can be seen?
- Mock: which owner kinds and request sizes the mock provider seeds (it currently seeds five uniform Running Pods), and whether the seed is fixed or randomised.
- Playground (`deploy/kind-chaos-playground`): today it is five uniform Deployments with no requests, so every Pod would be a Light Probe. Should it add a StatefulSet, a DaemonSet, a Job (and possibly a CronJob) and varied requests? Note that a DaemonSet on a single-node kind cluster yields only one Pod, and Job Pods complete, so the Jobs must keep running long enough to be shot.
- Does the playground need RBAC changes? (Research found no RBAC manifests in `deploy/`.)

## Answer

- **Mock seed for `--mock`**: a fixed, deterministic set of 9 eligible Pods plus one unmanaged Pod. The median is 100m/128Mi, so the tiers come out as listed.

  | Pod | Owner | Requests (CPU/mem) | Archetype / Tier | Symptoms |
  |---|---|---|---|---|
  | web-a | ReplicaSet | none | Probe / Light | – |
  | web-b | ReplicaSet | 100m/128Mi | Probe / Standard | – |
  | web-c | ReplicaSet | 500m/512Mi | Probe / Heavy | 2 restarts |
  | db-0 | StatefulSet | 100m/128Mi | Turret / Standard | – |
  | db-1 | StatefulSet | 500m/512Mi | Turret / Heavy | – |
  | agent-x | DaemonSet | 25m/32Mi | Sentinel / Light | – |
  | agent-y | DaemonSet | 100m/128Mi | Sentinel / Standard | not ready |
  | batch-1 | Job | 25m/32Mi | Dart / Light | – |
  | batch-2 | Job | 100m/128Mi | Dart / Standard | 6 restarts |
  | stray | none (unmanaged) | – | never shown | – |

- **Mock API**:
  - `NewMockProvider(count)` stays as it is for existing tests: uniform ReplicaSet Pods with no requests, so all are Light Probes.
  - Add `NewMockProviderWithPods(pods ...PodInfo)`. The seed above is exposed as `cluster.MockVarietyPods()`, which `--mock` uses.
  - Add `SetSymptoms(ns, name, ready, restarts)`, which emits Modified events for tests.
  - The mock sends `EventSynced` after its seed events, per [Shaping the trait payload](09-shaping-the-trait-payload.md).
- **Simulated live activity**: only when started by `--mock`, never in tests.
  - A ticker flips `agent-y` between ready and not-ready every ~10 s, and gives `web-c` one more restart every ~20 s.
  - Simulated controllers: ~2 s after a managed Pod's deletion completes, the mock adds a replacement with the same owner kind and requests, a new name suffix, 0 restarts, and not-ready for ~3 s. `stray` is never recreated. Tests can opt in to replacements.
- **Chaos playground** (`deploy/kind-chaos-playground`):
  - The five Deployments are kept, with requests of none, 10m/16Mi, 50m/64Mi (two of them) and 250m/256Mi.
  - Add a StatefulSet with 2 replicas at 50m/64Mi, giving Standard Turrets.
  - Add a DaemonSet at 10m/16Mi, giving one Light Sentinel on single-node kind.
  - Add a Job running `pause` with `parallelism: 2`, `completions: 2`, `backoffLimit: 100` and 250m/256Mi, giving Heavy Darts. Its Pods never complete on their own, and chaos deletions don't exhaust the retry limit.
  - The median is 50m, so every tier is represented.
  - No CronJob, because a Job already stands for it.
  - Keep the existing playground labels on everything.
- **RBAC**: no change. Owner kind comes from the Pod's own `ownerReferences`, so Pod list, watch and delete are still enough.
