# Practice session

Practice starts a session that spawns one enemy per eligible Pod. Killing enemies scores `Kills`
and leaves the Pods alone, so killed enemies come back after a delay.

## Sub-features

- `practice-start` spawns the seeded enemies around the centred ship, with HUD `Enemies: 9` and `Kills: 0` in `mock`.
- `practice-hud` shows the HP bar `HP n/10`, the `SPACE Blaster` / `X Rockets` panel and the `Z Roll` bar.
- `practice-fire` fires once per Space press (twin tracers), and X fires a Rocket.
- `practice-kill` raises `Kills: N` when an enemy dies.
- `practice-respawn` brings a killed enemy back, so `Enemies` returns to 9.
- `practice-steer` turns the ship with left/right and thrusts with up.

## How to get to it (user POV)

- Title → Enter → (cursor on Practice) → Enter.
- `R` on the game-over screen of a practice session (see game-over-and-retry).

## Driving it with podctl

Preconditions:

- A `mock` run is on the title screen.

- **Start practice.** Run `$P key <id> 80 enter; sleep 0.3; $P shot <id> mode-select; $P key <id> 80 enter; sleep 1; $P shot <id> practice-start`. The second frame shows the ship at centre, `HP 10/10`, `Enemies: 9`, `Kills: 0` and no `!! CHAOS MODE !!` banner.
- **Spray fire.** Run `$P spray <id> 100; $P shot <id> after-spray`. The ship has rotated, and the frame shows `Kills:` ≥ 1 and still `Enemies: 9` (the killed enemy is pending respawn or already back). The frame may already be `G A M E   O V E R` with `Kills: N`, which also proves the kill.
- **Rocket.** Run `$P key <id> 80 x; sleep 0.2; $P shot <id> rocket`. A missile is in flight and the Rockets bar is recharging.
- **Leave.** Let the ship die and follow game-over-and-retry, or `$P stop <id>`.

## Gotchas

- Standing still is lethal. Turret Bolts and lunges drain HP within ~10–20 s, so take shots right after the action.
- `spray 100` scored 2, 3 and 3 kills in three proof runs. A 60-tap Blaster-only spray once scored 0. Kills are probabilistic, so if `Kills: 0`, retry once before calling it a failure.
- `Enemies` counts the pool, including enemies pending respawn, so it stays at 9 in practice. A drop below 9 that persists is the bug signal.
- The `mock` ticker flips `agent-y`'s readiness every ~10 s and adds a `web-c` restart every ~20 s, so expect Symptoms to change mid-run.
