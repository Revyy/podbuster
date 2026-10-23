---
name: verify-podbuster
description: Launch the real Podbuster desktop game (Ebitengine window on macOS), drive it with keyboard events, and capture screenshots and logs as proof. Use when a change to game/, bridge/, cluster/ or main.go needs to be shown working in the running app, not just in go test — menus, practice play, chaos deletions against the kind playground, the Field guide.
---

# Verify Podbuster

Podbuster is a single desktop window (800×600, title `Podbuster 🚀`) with no API and no CLI
output beyond a log. You drive it by posting key events to the game's PID and read it back
through Ebitengine's built-in screenshot key. Everything goes through one helper:

```bash
P=.claude/skills/verify-podbuster/scripts/podctl.sh   # run from the repo root
```

| Command | What it does |
| --- | --- |
| `$P launch <run-id> [mock\|kind]` | Builds the game from the working tree, starts it, waits for the window, takes `01-ready.png`. |
| `$P doctor <run-id>` | Read-only health check. Any `FAIL` line means don't drive it. |
| `$P key <run-id> <holdMs> <key>...` | Presses the keys together, holds them, releases them. |
| `$P spray <run-id> [taps]` | Turns left while firing every tap, plus a Rocket every 20th and a roll every 45th (~12 s for 100). Scores kills without aiming. |
| `$P shot <run-id> <name>` | Saves the current frame to `shots/NN-<name>.png` and prints the path. |
| `$P stop <run-id>` | Kills this run's game process. Evidence is kept. |

Keys: `up down left right w a s d space x z enter escape backspace n p r 1 2 3 f12`.

## Launch

Prerequisites (macOS only): Go 1.26.3, `swiftc` (Xcode command line tools), and **Accessibility
permission for the terminal app hosting the agent** (System Settings → Privacy & Security →
Accessibility). Without that, macOS silently drops every key event. `launch` checks this and
refuses to start. Screen Recording permission is *not* needed.

```bash
$P launch fix-123 mock      # in-memory variety seed, namespace "verify"; no cluster
$P launch fix-123 kind      # kind playground; run `make kind-up` first (see Chaos)
```

`launch` compiles `scripts/keys.swift` into `.verify/bin/keys` if needed, then runs `go build`
into `.verify/runs/<run-id>/podbuster`. It starts the game with `EBITENGINE_SCREENSHOT_KEY=F12`
in that directory and waits for the window. Readiness is a rendered frame, not just a live
process: `launch` prints `ready: run <id> pid <pid> dir <dir>` only after `01-ready.png` was
written. The game starts on the title screen. Each launch brings the window to the front once.

There is deliberately no mode for a real cluster. Chaos mode deletes Pods, so verification
only ever points at `--mock` or the disposable kind playground.

Several runs can be live at once under different run ids. Each has its own process, binary and
directory, and keys are posted to one PID. A reused run id refuses to launch while its game is
still running.

## Doctor

```bash
$P doctor fix-123
```

This checks that the pid file names a live process running this run's binary, that the window is
on screen, that Accessibility is granted, and that the build revision equals `HEAD`. It also prints
`run.env` (mode, flags, rev, count of dirty Go files) and the last log lines. Run it first whenever
a key seems to do nothing or a shot fails. If the rev differs or `dirty` doesn't match your edits,
stop and relaunch.

## Drive

The game runs in real time between commands: enemies move, Turrets fire, and the player dies
in roughly 10–20 s of standing still. Chain the key presses and the shot in one shell
command when the frame matters.

- Menus react to *just pressed* keys: use `$P key <id> 80 enter`, then `sleep 0.3`.
- Weapons fire once per press: `space` is the Blaster, `x` Rockets, `z` a Barrel roll.
  Holding a key doesn't repeat fire, so use `$P spray` (or loop `$P key`, written so your shell
  splits the key names).
- Steering and thrust are *held*. `$P key <id> 600 up` thrusts for 600 ms, and
  `$P key <id> 70 left space` turns while firing one shot.
- Escape does nothing during play or Dying. To leave play, let the ship die (wait for
  `G A M E   O V E R`) and then press `escape`.
- Keys reach the game while it is in the background, so you don't need to focus it.

The Field guide, chaos journeys and so on are in [features/README.md](features/README.md). Read
the matching feature file before driving.

## Evidence

Everything for a run is in `.verify/runs/<run-id>/` (gitignored):

- `shots/NN-<name>.png` are frames in action order. Read them with the image viewer and quote the
  HUD text (`Enemies: N`, `Kills: N` / `Pods deleted: N`, `HP n/10`) that proves the claim.
- `game.log` holds the game's stderr: provider line, `bridge:` failures and stale-kill discards,
  `Saved screenshot:` lines.
- `run.env` holds the mode, flags and git rev the proof was taken against.

Proof standards:

- Drive the real key path from the title screen. Don't call game functions, inject `GameEvent`s,
  or use test helpers. Those are what `go test` already covers.
- Shoot before *and* after the action. A single final frame doesn't show that your input caused it.
- For chaos, the visible counter is not enough. Prove the side effect with `kubectl` against the
  kind playground (see [features/chaos-session.md](features/chaos-session.md)).
- `--mock` is the production `MockProvider` seed and the sanctioned stand-in for a cluster. Say
  which mode a proof used.

## Cleanup

```bash
$P stop fix-123               # kills only the PID this run recorded, deletes its binary
make kind-down                # only if this verification created the kind cluster
```

`stop` never kills by process name, and it leaves `shots/`, `game.log` and `run.env` in place.
Delete `.verify/runs/<run-id>/` yourself once the proof has been reported. After a failed attempt,
run `stop` before relaunching, or the stale process will keep running.

## Helpers

- `scripts/podctl.sh` is the entry point above.
- `scripts/keys.swift` is the CoreGraphics key poster. `keys trusted`, `keys window <pid>`,
  `keys activate <pid>` and `keys tap <pid> <holdMs> <key>...` all work directly once built
  (`.verify/bin/keys`). Add a key by extending its `codes` table with the macOS virtual keycode.
- `PODBUSTER_VERIFY_ROOT` moves `.verify/`, and `KIND_CLUSTER_NAME` / `PODBUSTER_NAMESPACE`
  match the Makefile overrides for `kind` mode.
