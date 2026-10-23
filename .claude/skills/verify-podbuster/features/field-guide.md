# Field guide

The Field guide is a paged, non-mode screen reached from mode select. It explains the four enemy
archetypes with live demos, Tiers & Bolts, Symptoms, and an illustrative `kubectl`-style table of
how Pods become enemies. It reads no cluster data.

## Sub-features

- `guide-open` opens on the Probe page with Enter on `FIELD GUIDE`.
- `guide-page` pages with right/d and left/a, wrapping across seven pages: Probe, Turret, Sentinel, Dart, Tiers & Bolts, Symptoms and Pods.
- `guide-tier` shows the archetype demo at Light, Standard or Heavy with `1`/`2`/`3`.
- `guide-symptoms` moves rows with up/down on the Symptoms page, toggles not ready with `n`, cycles restart plates with `p` and starts a reboot blink with `r`.
- `guide-leave` returns to mode select with Escape or Backspace, with the cursor on `FIELD GUIDE`.

## How to get to it (user POV)

- Title → Enter → down → down → Enter.

## Driving it with podctl

Preconditions:

- A `mock` run is on the title screen.

- **Open.** Run `$P key <id> 80 enter; sleep 0.3; $P key <id> 80 down; sleep 0.3; $P key <id> 80 down; sleep 0.3; $P shot <id> cursor-guide; $P key <id> 80 enter; sleep 0.5; $P shot <id> guide-probe`. The first frame has `>` on `FIELD GUIDE`, and the second shows the Probe page with a moving demo.
- **Page.** Run `$P key <id> 80 right; sleep 0.5; $P shot <id> guide-turret`. The Turret page shows.
- **Tier.** Run `$P key <id> 80 3; sleep 0.5; $P shot <id> guide-heavy`. The demo enemy is the Heavy variant.
- **Symptoms.** Press `right` until the Symptoms page shows (four more presses from Turret), then run `$P key <id> 80 n; $P key <id> 80 p; sleep 0.3; $P shot <id> symptoms`. The demo shows the not-ready eye and patch plates.
- **Leave.** Run `$P key <id> 80 backspace; sleep 0.3; $P shot <id> back`. Mode select shows with `>` on `FIELD GUIDE`.

## Gotchas

- The guide opens no watch. In `kind` mode `game.log` must show no new `bridge:` lines while in the guide.
- Escape in the guide goes to mode select, not the title. A second Escape is needed for the title.
- Page titles and layout are drawn art. Check the frame visually rather than expecting exact strings.
