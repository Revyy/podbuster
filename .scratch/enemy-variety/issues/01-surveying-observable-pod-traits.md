# Surveying observable Pod traits

Type: research
Status: resolved
Blocked by:

## Question

Which Pod characteristics can Podbuster observe cheaply and reliably through its existing client-go namespace informer (`cluster/kube.go`)? Which of them are worth using for Archetype or Tier?

Evaluate each candidate: owner kind (ReplicaSet→Deployment resolution, StatefulSet, DaemonSet, Job/CronJob, bare Pod), CPU/memory requests and limits (including Pods with none), QoS class, restart count, container/sidecar/init-container count, age, readiness, and priority class.

For each, report:
- where it lives on the Pod object
- whether extra API calls, informers, or RBAC verbs are needed (e.g. resolving a ReplicaSet's owning Deployment)
- whether it is fixed or changes over a Pod's life
- how it is typically distributed in real namespaces, i.e. whether it would actually vary between enemies

## Answer

- Podbuster's Pod informer already sees full `corev1.Pod` objects; current `PodInfo` only drops most fields.
- Best Archetype source: controller owner kind from `metadata.ownerReferences[controller=true]`.
- Direct owner kind is free; ReplicaSet→Deployment and Job→CronJob resolution needs extra owner reads/informers/RBAC.
- Best Tier source: effective CPU/memory requests from `spec.containers[].resources.requests`, banded relative to the namespace.
- Pods with no requests should remain the lowest tier; limits are secondary/fallback because requests describe scheduling reservation.
- QoS (`status.qosClass`) is cheap and reliable but only a coarse derivation of requests/limits.
- Readiness and restart count are cheap live signals, but should be cosmetic overlays because they change after spawn.
- Container/init/sidecar counts, age, and priority class are observable but weak or sparse for first-pass gameplay mapping.
- Existing `deploy/` has no RBAC manifests; current playground targets are five uniform Deployments with no requests/limits.
- Shortlist: owner kind for Archetype; resource requests for Tier; QoS fallback; readiness/restarts as overlays.

Full findings: branch `research/observable-pod-traits`, file `docs/research/observable-pod-traits.md`.

