Type: task
Status: resolved

## Question

`tdd/tests.md` illustrates good vs. bad tests with TypeScript/Jest snippets (a `checkout`/cart
example, `jest.mock`, a raw SQL assertion against a `users` table).

Replace these with Go snippets. This repo's test suite (`go test ./...`) already has real
examples to draw from — check `bridge/` or `cluster/mock.go` tests for an existing
integration-style test using `MockProvider`, and adapt or reference that pattern for each
category:

- **Good / integration-style**: a Go test that drives real interfaces (e.g. spawning a
  `MockProvider`, adding a Pod, asserting on the resulting `PodEvent` or game-side enemy spawn)
  rather than mocking internals.
- **Bad / implementation-detail test**: a Go test that reaches into an unexported field or asserts
  call counts on an internal collaborator instead of observable behavior.
- **Bad / bypasses interface to verify**: a Go example that inspects internal state directly
  instead of through the public method, paired with the "good" version that verifies through the
  interface.
- **Tautological test**: a Go example where the expected value is recomputed the same way the
  code computes it, versus one with an independent literal.

Resolve by rewriting `.agents/skills/tdd/tests.md` in place, keeping its structure (Good Tests /
Bad Tests / red flags list / Tautological tests) and the "Characteristics" / "Red flags" bullet
lists unchanged except where they name TS/Jest-specific concepts (e.g. `jest.mock` → Go's manual
fake/interface substitution).

## Answer

Rewrote `.agents/skills/tdd/tests.md`'s four code examples in Go, drawing on the repo's real
`bridge/bridge_test.go` pattern (`MockProvider` + `startSession`/`expectEvent` helpers):

- **Good/integration-style**: adapted `TestBridge_SpawnEnemyOnPodAdded` directly — real
  `MockProvider`, real `Bridge`, asserting on the resulting `game.EventSpawnEnemy`.
- **Bad/implementation-detail**: a hypothetical spy counting `Watch` calls on a faked provider.
- **Bad/bypasses interface**: reaching into `MockProvider`'s unexported `pods` map, vs. **good**:
  verifying via `ListPods`.
- **Tautological**: a `CalculateTotal` example recomputing the sum the same way the code does, vs.
  an independent literal (`15`).

Kept the "Characteristics" / "Red flags" bullet lists as-is except swapping `jest.mock` language
for "faking internal collaborators." Verified event/method names (`game.EventSpawnEnemy`,
`MockProvider.AddPod`) against the actual source before committing the snippets.
