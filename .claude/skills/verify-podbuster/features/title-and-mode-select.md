# Title and mode select

The title screen leads to a mode select that lists Practice, Chaos and, set apart below a rule,
the Field guide. It shows the watched namespace. Chaos costs a second Enter and can be backed out
of with Escape.

## Sub-features

- `title-enter` opens mode select with Enter or Space, cursor on Practice.
- `mode-cursor` moves the `>` marker with up/down or w/s, wrapping through all three entries.
- `mode-namespace` shows `Namespace: <ns>  (set with --namespace)`.
- `chaos-confirm` shows a DELETE warning on the first Enter on Chaos. Only a second Enter starts it.
- `chaos-backout` clears the warning with Escape and keeps the cursor on Chaos. The cursor is frozen while the warning shows.
- `mode-escape` returns to the title with Escape when no warning is showing.

## How to get to it (user POV)

- Press Enter (or Space) on the title screen.
- Press Escape or Backspace in the Field guide (cursor lands on Field guide).
- A watch-start failure returns here with `Could not watch namespace: …` at the bottom.

## Driving it with podctl

Preconditions:

- A `mock` run is on the title screen.

- **Open mode select.** Run `$P key <id> 80 enter; sleep 0.3; $P shot <id> mode-select`. The frame shows `SELECT MODE`, `Namespace: verify` and `> PRACTICE`.
- **Move the cursor.** Run `$P key <id> 80 down; sleep 0.3; $P shot <id> cursor-chaos`. The `>` sits on `CHAOS`.
- **Arm chaos.** Run `$P key <id> 80 enter; sleep 0.3; $P shot <id> chaos-warning`. The frame shows `This will DELETE real Pods in namespace verify.` and `Press ENTER again to confirm, ESC to back out.`
- **Cursor frozen.** Run `$P key <id> 80 down; sleep 0.3; $P shot <id> frozen`. The warning remains and the `>` is still on `CHAOS`.
- **Back out.** Run `$P key <id> 80 escape; sleep 0.3; $P shot <id> backed-out`. The warning is gone, the hint line `UP/DOWN to choose…` is back, and the cursor is on `CHAOS`.
- **Back to title.** Run `$P key <id> 80 escape; sleep 0.3; $P shot <id> title`. The title screen shows.

## Gotchas

- In `mock` mode a confirmed chaos start "deletes" Mock Pods only, but never press the second Enter in a `kind` run unless the chaos journey is the one under test.
- Each frame must be taken after the 0.3 s settle. Earlier frames can predate the key.
- The title screen has no Field guide hint. That is by design and not a missing feature.
