Type: task
Status: resolved

## Question

`tdd/mocking.md` illustrates "mock at system boundaries" and "design for mockability" (dependency
injection, SDK-style interfaces over generic fetchers) with TypeScript snippets (`processPayment`,
a Stripe client, a generic `fetch` API object).

Replace these with Go snippets using this repo's own boundary: `cluster.PodProvider` (the real
system boundary between `bridge/` and the cluster, satisfied by `MockProvider` or `KubeClient`).
Show:

- The "easy to mock" / "hard to mock" DI contrast using a Go function or struct that accepts a
  `PodProvider` versus one that constructs a `KubeClient` internally.
- The "SDK-style interface over generic fetcher" contrast using `PodProvider`'s three methods
  (`ListPods`, `Watch`, `DeletePod`) versus a hypothetical generic `Do(op string, args ...any)`-style
  interface.

Keep the "When to Mock" / "Don't mock" bullet list's guidance as-is (it's language-agnostic);
only replace the code blocks and any TypeScript-specific phrasing ("classes/modules" is fine,
already generic). Resolve by rewriting `.agents/skills/tdd/mocking.md` in place.

## Answer

Rewrote `.agents/skills/tdd/mocking.md`'s two code snippets in Go, using this repo's real
`cluster.PodProvider` boundary:

- **DI contrast**: `newBridge(provider cluster.PodProvider)` (easy to mock — swap in
  `cluster.NewMockProvider(5)`) versus a constructor that builds `cluster.NewKubeClient(...)`
  internally (hard to mock).
- **SDK-style contrast**: the real `PodProvider` interface (`ListPods`/`Watch`/`DeletePod`, each
  independently mockable) versus a hypothetical generic `Do(op string, args ...any)` interface
  requiring conditional logic in the fake.

The "When to Mock" / "Don't mock" bullet lists and the "SDK approach means" bullets were left
as-is (already language-agnostic); only code blocks and captions were replaced.
