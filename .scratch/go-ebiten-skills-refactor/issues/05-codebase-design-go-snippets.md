Type: task
Status: resolved

## Question

`codebase-design/SKILL.md` has three TypeScript code snippets under "Designing for testability"
(`processOrder(order, paymentGateway)` vs. constructing a `StripeGateway` internally;
`calculateDiscount` returning a value vs. `applyDiscount` mutating `cart.total`) and one prose
reference to "the TypeScript `interface` keyword" in the Rejected framings section.

Replace the three code snippets with Go equivalents. Prefer this repo's own module boundaries as
the concrete example, e.g.:

- Accept-dependencies vs. construct-internally: a Go function/constructor taking a
  `cluster.PodProvider` parameter versus one that calls `cluster.NewKubeClient(...)` internally.
- Return-results vs. side-effects: a pure function computing a value (e.g. deriving eligible pod
  count, or a game scoring/damage calculation) versus one that mutates shared state in place.

Update the Rejected framings bullet ("'Interface' as the TypeScript `interface` keyword...") to
reference Go's `interface` keyword or a struct's exported method set instead, keeping the same
point (interface-as-used-here is broader than the language keyword). Resolve by rewriting
`.agents/skills/codebase-design/SKILL.md` in place; leave the glossary, ASCII diagrams, and
Relationships/Rejected-framings structure otherwise unchanged.

## Answer

Rewrote `.agents/skills/codebase-design/SKILL.md`'s three code snippets in Go:

- **Accept-dependencies vs. construct-internally**: `processKill(enemy, provider
  cluster.PodProvider)` vs. one that calls `cluster.NewKubeClient(...)` internally.
- **Return-results vs. side-effects**: `calculateDamage(hit HitInfo) int` (pure) vs.
  `applyDamage(enemy *Enemy, hit HitInfo)` (mutates `enemy.health` in place).
- Updated the Rejected-framings bullet to reference Go's `interface` keyword and a struct's
  exported method set instead of TypeScript's `interface` keyword, keeping the same point.

Glossary, ASCII diagrams, and the Relationships section were left untouched — no code or
TS-specific language there.
