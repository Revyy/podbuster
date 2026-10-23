# 12: Bridge-owned eligibility and synced watch

**What to build:** Prefactor. Eligibility becomes a single bridge rule, and the watch signals when its initial sync has finished. The rule is that a Pod is a managed Pod (it has a controller owner) and is not in a terminal phase. Providers report every Pod raw. See decision [Shaping the trait payload](09-shaping-the-trait-payload.md).

**Blocked by:** None (can start immediately)

**Status:** resolved

Spec: [spec.md](../spec.md)

- [x] `PodInfo` reports the controller owner kind; an empty kind means unmanaged. The real provider and the mock both fill it in. The mock's default Pods are owned by a ReplicaSet.
- [x] The real provider no longer filters terminal Pods or turns them into deletions. The bridge applies one eligibility function to every event.
- [x] A known Pod that becomes ineligible despawns through the existing deletion path. An unknown Pod that becomes eligible on a Modified event spawns.
- [x] Unmanaged Pods never spawn.
- [x] The watch emits a synced marker. The real provider's `Watch` returns immediately and sends the marker once the event handler has synced. The mock sends it after its seed events.
- [x] A test with more than 100 Pods on a fake clientset passes without blocking.
- [x] The bridge holds initial spawns until the marker, then emits them.
- [x] Bridge and KubeClient tests cover all of the above. `go test ./...` is green.

## Completion

- **Completed:** 2026-10-01
- **Commit:** `853ab7c`, `a847992` (merged into `feat/enemy-variety-impl`)
- **Notes:** A Deleted event for an unknown Pod no longer despawns. A Pod already Terminating when first seen spawns, then gets a terminating event at sync. A Pod re-added before sync spawns once.
