# Game over and retry

When HP reaches 0 the ship breaks apart (Dying). Then the game-over screen summarises the
session. `R` retries in the same mode with the same session and enemies, and Escape ends the
session and returns to the title.

## Sub-features

- `dying` breaks the ship into spinning pieces with a fireball. Input is ignored.
- `gameover-summary` shows `G A M E   O V E R`, `Kills: N` (or `Pods deleted: N` in chaos), `R to retry in the same mode` and `ESC to return to the title screen`.
- `retry` restores `HP 10/10` with the ship at centre, keeps the enemy pool and keeps the kill count.
- `gameover-escape` ends the session and shows the title screen.
- `chaos-retry-notice` adds `Chaos mode is still active on retry.` in chaos.

## How to get to it (user POV)

- Lose all HP during a practice or chaos session.

## Driving it with podctl

Preconditions:

- A practice session is running (see practice-session).

- **Die.** Stop pressing keys and run `sleep 20; $P shot <id> game-over`. The frame shows `G A M E   O V E R`, `Kills: N` and both hint lines, and the HUD shows `HP 0/10`.
- **Retry.** Run `$P key <id> 80 r; sleep 0.5; $P shot <id> retried`. The frame shows `HP 10/10`, the ship at centre, `Kills:` unchanged and the same `Enemies` count.
- **Escape to title.** Die again (`sleep 20`), then run `$P key <id> 80 escape; sleep 0.3; $P shot <id> title`. The title screen shows.

## Gotchas

- Escape during Dying or play is ignored. Wait until the game-over text is on the frame.
- Death time varies with where Turrets and Darts are. Check for `G A M E   O V E R` and wait longer if it isn't there yet, rather than pressing keys.
- Retry is `R`, the same key the Field guide uses for a reboot blink. Only the game-over screen treats it as retry.
