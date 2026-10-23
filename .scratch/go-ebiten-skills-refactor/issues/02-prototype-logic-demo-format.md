Type: grilling
Status: resolved

## Question

`prototype/LOGIC.md` describes a single self-contained HTML file (plain HTML/CSS/JS, no
framework, no bundler, no server) that lets a non-developer drive a state model by clicking
buttons and reading rendered state, with free-play buttons and guided walkthrough tabs.

Decide whether this stays HTML/JS or becomes Go-native for this repo:

- **Keep HTML/JS**: the pure module (reducer/state machine) gets reimplemented in vanilla JS
  purely for the demo; it's a translation exercise, not the real Go code, but it's still the most
  "double-click and share with anyone, no install" format available, and the LOGIC.md skill
  already treats the demo as disposable (the validated design "lifts into the real module" by
  hand afterward, not by copy-paste).
- **Go-native**: e.g. a single `go run <file>.go` terminal program (still no build step required
  beyond having Go installed, still shareable as one file) that prints state after each action and
  offers numbered menu choices instead of clicking buttons — trading "double-click, zero
  install" for "matches the real language, no reimplementation-in-JS translation risk."
- Consider: this repo's target audience for a logic prototype (the player-facing state machines
  in `game/`, e.g. session lifecycle, enemy respawn logic) — is a non-developer really the
  audience, or is it always another engineer, which would favor Go-native?

Resolve by rewriting `.agents/skills/prototype/LOGIC.md` in place with the chosen format,
preserving its structure (When this is the right shape / Process steps 1–5 / Anti-patterns) and,
if switching away from HTML, dropping HTML-specific advice (e.g. "no `<script>` framework",
"double-clicks to open") for the Go equivalent phrasing. Update the one-line branch description in
`.agents/skills/prototype/SKILL.md` to match.

## Answer

Rewrote `.agents/skills/prototype/LOGIC.md` for Ebitengine/Go, going fully Go-native:

- **Dropped the "non-developer" framing.** This repo's audience for any prototype is always
  another engineer; matching the real language removes reimplementation-in-JS translation risk.
- **Format**: a single `//go:build prototype`-tagged Go file, run with
  `go run -tags prototype <path>` — consistent with UI.md's build-tag convention.
- **Interaction shape**: replaces free-play buttons with a numbered menu of legal actions printed
  after every state change; replaces tabbed guided walkthroughs with named scenarios selectable
  from an initial menu, which auto-play their scripted steps (with a short pause and description
  per step) then drop into free-play from wherever they end.
- **Logic isolation**: kept the same shape guidance (pure reducer / state machine / pure functions
  / struct with method surface) translated to Go idioms, with the same "no I/O inside the pure
  package" discipline — now mirroring this repo's own `game.State` machine as the state-machine
  example.
- Capture step: the validated logic lifts into the real package "usually needing only import-path
  and naming changes since it was already Go" (no cross-language port required, unlike the old
  HTML/JS version).

Also updated `.agents/skills/prototype/SKILL.md`'s LOGIC branch one-liner and rule 2 (trivial to
run) to describe the `go run -tags prototype` invocation instead of "a single HTML file the user
double-clicks."
