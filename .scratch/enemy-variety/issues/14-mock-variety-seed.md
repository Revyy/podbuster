# 14: Mock variety seed

**What to build:** `--mock` shows every archetype and tier, and simulates controllers recreating deleted Pods. See decision [Choosing mock and playground variety](10-choosing-mock-and-playground-variety.md). The Symptom fields and ticker belong to ticket 19.

**Blocked by:** 13

**Status:** resolved

Spec: [spec.md](../spec.md)

- [x] The mock has a constructor that takes explicit Pods. A function exposes the representative 9+1 seed table from decision 10. `--mock` uses it. `NewMockProvider(count)` is unchanged for existing tests.
- [x] `stray`, the unmanaged Pod, never appears.
- [x] An opt-in replacement behaviour, enabled by `--mock`: ~2 s after a managed Pod's deletion completes, a replacement with the same owner kind and requests and a new name suffix is added. `stray` is never recreated.
- [x] A test asserts the seed's resulting Archetype and Tier mix through the bridge.
- [x] `go test ./...` is green.

## Completion

- **Completed:** 2026-10-01
- **Commit:** `e20493f` (merged into `feat/enemy-variety-impl`)
- **Notes:** Adds NewMockProviderWithPods, MockVarietyPods and an opt-in EnableReplacements(delay); --mock uses a 2 s delay. The seed lives in namespace default, as the old mock did.
