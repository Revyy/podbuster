# Go/Ebitengine skills refactor

## Destination

Every skill file under `.agents/skills/` that currently illustrates its process with
TypeScript/web-development examples — `prototype`'s UI and LOGIC branches, `tdd`'s mocking and
tests snippets, `codebase-design`'s snippets, `diagnosing-bugs`' browser-automation step —
instead illustrates it with Go and Ebitengine-native examples that match this repo's actual
stack. The underlying process guidance (what each skill teaches) does not change, only the
examples it teaches through.

## Notes

- **Override: this map executes, not just decides.** Each ticket both decides the replacement
  content and edits the file(s) in place. This is a content rewrite, not a build with a separate
  execution phase.
- Domain: this repo (Podbuster, a Go/Ebitengine desktop game; see root `AGENTS.md` and
  `CONTEXT.md`). Prefer this repo's own real code (`cluster.PodProvider`, `MockProvider`,
  `KubeClient`, `game/`, `bridge/`) as the concrete Go example wherever a ticket needs one, so
  replacements feel native rather than generic.
- For the two genuine design-decision tickets (prototype UI, prototype LOGIC), call the Skill
  tool for "grilling" and "domain-modeling" as needed — these have no single obviously-correct
  Go/Ebiten equivalent and are worth a real conversation.
- No research tickets: this is a repo-internal content task, no external facts required.

## Decisions so far

- [Prototype UI variant switching](issues/01-prototype-ui-variant-switching.md): variants switch on-screen via a `Tab`-cycled `variant int` field and `ebitenutil.DebugPrint` overlay, all gated behind a `//go:build prototype` tag; sub-shape B is a standalone `game/prototype/<name>/main.go` entry point.
- [Prototype LOGIC demo format](issues/02-prototype-logic-demo-format.md): dropped HTML/JS and the non-developer framing entirely; the demo is now a `//go:build prototype`-tagged `go run` terminal program with a numbered free-play menu plus auto-playing named scenarios.
- [TDD mocking Go example](issues/03-tdd-mocking-go-example.md): replaced the TS Stripe-client snippets with this repo's real `cluster.PodProvider` boundary (DI via `MockProvider`/`KubeClient`; SDK-style `ListPods`/`Watch`/`DeletePod` vs. a generic `Do(op, args...)`).
- [TDD tests Go example](issues/04-tdd-tests-go-example.md): replaced the TS/Jest checkout snippets with Go examples built on the real `bridge_test.go` pattern (`MockProvider` + `EventSpawnEnemy`) for good/bad/tautological cases.
- [Codebase-design Go snippets](issues/05-codebase-design-go-snippets.md): replaced the three TS testability snippets with Go (`PodProvider` DI, pure `calculateDamage` vs. mutating `applyDamage`) and swapped the Rejected-framings TS `interface` reference for Go's.
- [Diagnosing-bugs Ebiten loop](issues/06-diagnosing-bugs-ebiten-loop.md): replaced the Playwright/Puppeteer headless-browser list item with a headless `game.NewGame` + `Update()` harness, with pixel-inspection as a fallback for visual bugs.

## Not yet specified

(none — the frontier covers the full scope; nothing further is expected to graduate)

## Out of scope

- `teach`'s HTML-lesson format: HTML lessons are a deliverable format for humans (print-friendly reference docs, interactive lessons), not a stand-in "web dev" example needing a Go/Ebiten equivalent.
- `setup-matt-pocock-skills`' `pnpm-workspace.yaml` check: real repo-detection logic for JS monorepos, not an illustrative example.
- `improve-codebase-architecture` (+ `HTML-REPORT.md`): the HTML report is a legitimate cross-platform deliverable format, orthogonal to the target language; it contains no source-language code snippets to swap.
- `wayfinder`, `unslop`, `writing-for-agents`: passing mentions of "UI"/"React"/"TypeScript" only, not examples that need replacing.
