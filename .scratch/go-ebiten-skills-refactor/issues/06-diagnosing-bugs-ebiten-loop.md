Type: task
Status: resolved

## Question

`diagnosing-bugs/SKILL.md`'s Phase 1 feedback-loop list includes, as option 4: "Headless browser
script (Playwright / Puppeteer) that drives the UI and asserts on DOM/console/network." This has
no meaning for a desktop Ebitengine game (no DOM, no browser, no console/network to assert on in
that sense).

Decide and write the Ebitengine-appropriate equivalent for this list entry: driving the actual
game loop headlessly and asserting on game state, e.g.:

- Constructing a `game.Game` (or the relevant scene/state) directly in a Go test, calling `Update()`
  N times with scripted input, and asserting on resulting state (enemy count, player health,
  score) — no `Draw()`/real window required since Ebitengine's `Update` doesn't need a live
  display for this kind of harness.
- Optionally mention capturing a rendered frame via `ebiten.Image` / `image.RGBA` pixel
  inspection for cases where the bug is visual, as a heavier fallback.

Keep the rest of the ordered list (failing test, CLI invocation, replay a captured trace,
throwaway harness, property/fuzz loop, bisection harness, differential loop, HITL bash script)
unchanged — they're already language/platform-agnostic. Resolve by rewriting the single list item
in `.agents/skills/diagnosing-bugs/SKILL.md` in place.

## Answer

Replaced the "Headless browser script" list item in `.agents/skills/diagnosing-bugs/SKILL.md`'s
Phase 1 feedback-loop list with a "Headless game-loop harness": construct `game.NewGame(...)`
directly in a Go test, feed scripted input/events over its channels, call `Update()` repeatedly,
and assert on resulting state — no `Draw()` or live window needed. Mentioned rendering to an
`*ebiten.Image` and inspecting pixels as a heavier fallback for visual bugs. Rest of the ordered
list (failing test, CLI invocation, replay trace, throwaway harness, property/fuzz loop, bisection,
differential loop, HITL bash script) left unchanged.
