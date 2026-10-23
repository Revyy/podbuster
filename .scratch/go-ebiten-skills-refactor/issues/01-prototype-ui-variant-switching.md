Type: grilling
Status: resolved

## Question

`prototype/UI.md` describes generating several radically different UI variations on a single
web route, switchable via a `?variant=` URL search param and a floating bottom bar (arrow keys,
`router.replace`, hidden in production via `NODE_ENV`). None of this maps directly onto a desktop
Ebitengine game: there is no route, no URL, and no separate "production build" the prototype needs
to be hidden from at runtime in the same way.

Decide the Go/Ebitengine-native replacement for this whole flow:

- What does "sub-shape A: adjustment to an existing page" become for a game? (e.g. rendering
  variants on top of / instead of an existing `game/scene.go` draw path, gated by a flag or key,
  with real game state feeding it — analogous to "real header, real sidebar, real data".)
- What does "sub-shape B: a new page" become? (e.g. a standalone `go run` entry point / build tag
  that boots straight into the variant switcher, analogous to a throwaway route.)
- What replaces the `?variant=` URL param as the switching mechanism? (e.g. a keypress cycling
  through variants each frame, an on-screen debug HUD.)
- What replaces "hidden in production via `NODE_ENV`"? (e.g. a build tag like `//go:build
  prototype`, so a stray prototype can't ship in a normal `go build`.)
- Keep the same shape of guidance (radically-different variants, N capped at 3–5, a shared
  switcher component/pattern, capture-and-clean-up at the end) — only the mechanics change.

Resolve by rewriting `.agents/skills/prototype/UI.md` in place with the Go/Ebitengine equivalents,
preserving its structure (When this is the right shape / Two sub-shapes / Process / Anti-patterns).
Also update the one-line branch description in `.agents/skills/prototype/SKILL.md` if the mechanism
it names (URL search param / floating bottom bar) changes.

## Answer

Rewrote `.agents/skills/prototype/UI.md` for Ebitengine:

- **Sub-shape A** (adjustment to an existing screen, preferred): variants render from the same
  draw call (e.g. `drawHUD`), gated by a `variant int` field on `Game`, switched via a `switch`
  statement, all behind a `//go:build prototype` tag.
- **Sub-shape B** (new entry point, last resort): a standalone `main` package under
  `game/prototype/<name>/main.go`, also `//go:build prototype`-tagged, run via
  `go run -tags prototype ./game/prototype/<name>`. Never wired into the real `GameState` machine.
- **Switching mechanism**: `Tab` key (arrow keys are claimed by player movement) cycles forward
  through variants each `Update()`, wrapping around.
- **Debug overlay**: replaces the floating bottom bar — plain `ebitenutil.DebugPrint` text,
  top-left, e.g. `variant: B (Sidebar layout) — [Tab] next`.
- **Hidden-in-production mechanism**: `//go:build prototype` build tag on the switch, cycler, and
  overlay files, so a plain `go build`/`go run .` never compiles prototype code in — chosen over
  an always-compiled `--prototype` CLI flag to avoid shipping dead code/flags permanently.

Also updated `.agents/skills/prototype/SKILL.md`'s UI branch one-liner and rules 1–2 to describe
the keypress/build-tag mechanism instead of URL routes.
