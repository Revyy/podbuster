# Podbuster — Agent Guide

## What This Is

Podbuster is a desktop space shooter written in Go with Ebitengine. Each enemy represents one
eligible Kubernetes Pod in a watched namespace. In chaos mode, killing an enemy requests deletion
of its Pod; that Kubernetes operation can fail.

Providers report every Pod with its raw traits (phase, terminating, owner kind, Ready, Restarts).
The bridge applies the single eligibility rule: a Pod is eligible when it is managed (has a
controller owner) and non-terminal (not Succeeded or Failed). A terminating Pod is still eligible
and stays represented until final removal.

## Commands

```bash
# Development; no cluster required.
go run . --mock

# Watch a real cluster.
go run . --namespace=default

# Run the test suite.
go test ./...

# Validate before opening a PR. CI runs these same targets.
make check   # make lint + make test
make lint    # go mod tidy -diff, golangci-lint fmt --diff, golangci-lint run
make test    # go build ./... and go test -race ./...
make fmt     # apply go mod tidy, formatters and golangci-lint autofixes
```

Run `make check` and get it green before creating a PR. `make lint` installs the pinned
golangci-lint version into `.bin/` on first use. The linter set lives in `.golangci.yml`, which
also enforces convention 2 through `depguard`. On Linux, building Ebitengine needs the X11 and GL
headers that `.github/workflows/ci.yml` installs.

The module requires Go 1.26.3. Chaos mode is deliberately not a CLI flag: the player selects it
in-game and confirms it with a second input. See
[ADR 0001](docs/adr/0001-chaos-mode-is-an-in-game-confirmed-choice.md).

## Architecture

```
┌──────────────────────────────────────┐
│                game/                 │
│ Ebitengine state, combat, rendering  │
│ Never imports cluster/ or bridge/    │
└────────────────┬─────────────────────┘
      GameEvent / KillEvent / SessionEvent channels
┌────────────────▼─────────────────────┐
│               bridge/                │
│ PodEvents → GameEvents               │
│ KillEvents → optional Pod deletion   │
└────────────────┬─────────────────────┘
             PodProvider interface
┌────────────────▼─────────────────────┐
│              cluster/                │
│ MockProvider or KubeClient           │
└──────────────────────────────────────┘
```

All cross-layer communication uses the typed channels defined by `game/events.go`. `Bridge.Run`
is goroutine-safe and idles when there is no active game session, so it reads no cluster data on
the title or mode-select screens.

## Package Responsibilities

### `cluster/`

- `provider.go` defines `PodProvider`, `PodInfo`, and Pod watch events.
- `mock.go` provides the in-memory development and test provider. It seeds named Running Pods
  (`NewMockProvider(n)` for uniform ones, `NewMockProviderWithPods` for explicit ones;
  `MockVarietyPods(namespace)` is the `--mock` seed, placed in the `--namespace` covering every Archetype and Tier plus an unmanaged
  `stray`), exposes `AddPod`, `RemovePod`, and `TerminatePod`, and can opt in to simulated
  controller replacements with `EnableReplacements` (replacements boot not ready for ~3 s). Its
  `SetSymptoms` emits Modified events and `SimulateActivity` is the `--mock` ticker (it targets the given namespace);
  map iteration means its initial event order is unspecified.
- `kube.go` provides `KubeClient`, backed by `client-go`. It loads kubeconfig/context options,
  watches a namespace using an informer, and deletes Pods. It reports every Pod raw, including
  terminal and unmanaged ones, with the controller owner kind in `PodInfo.OwnerKind` (empty means
  unmanaged), plus `Ready` (Ready condition) and `Restarts` (regular and sidecar containers).
  `Watch` returns its channel immediately and sends `EventSynced` once the initial
  listing has been delivered; the mock sends it after its seed events.

```go
type PodProvider interface {
    ListPods(namespace string) ([]PodInfo, error)
    Watch(ctx context.Context, namespace string) (<-chan PodEvent, error)
    DeletePod(ctx context.Context, namespace, name string) error
}
```

To add a data source, implement `PodProvider` and wire construction into `main.go`; game code
should not need to know the source changed.

### `bridge/`

`Bridge` opens one namespace watch for each `SessionEvent{Type: EventSessionStart}` and cancels it
on `EventSessionEnd`. It translates Pod added/deleted events to enemy spawn/despawn events and
terminating modifications to `EventPodTerminating`. A Modified event that changes a known Pod's Ready
or Restarts emits `EventSymptomsChanged`; other modifications are ignored.

The bridge owns eligibility: one `eligible` rule (managed, non-terminal phase) is applied to every
event. A known Pod that becomes ineligible despawns, an unknown Pod that becomes eligible on a
Modified event spawns, and unmanaged Pods never spawn. Initial spawns are held until `EventSynced`.

Spawn events carry an Archetype (from the controller owner kind: StatefulSet Turret, DaemonSet
Sentinel, Job Dart, anything else Probe) and a Tier (Light, Standard or Heavy). The Tier compares the
Pod's CPU and memory requests with the median of the session's known eligible Pods that have
requests. Initial Pods are banded together once the watch has synced, and later Pods on arrival.
A Pod with no requests is Light. Both are fixed at spawn. Spawns also carry the Pod's Symptoms
(`Ready`, `Restarts`), which are live and cosmetic only.

A failure to start a watch is logged and emitted as `EventError`, which ends the game session.
If an already-open provider event channel closes, the bridge cancels it and returns to idle without
emitting an error. The bridge logs and drops game events when its outbound buffer is full.

In chaos mode, a `KillEvent` causes `DeletePod` only when its session ID matches the active
session. This prevents a buffered practice-mode or stale-session kill from deleting a Pod in a
later chaos session. Deletion failures are logged and emitted as `EventDeletionFailed` so the game
can revert its optimistic termination state.

### `game/`

- `events.go` owns all event types shared with the bridge, including session IDs. Spawn events
  carry Archetype, Tier and Symptoms; `EventSymptomsChanged` updates a live enemy's Symptoms.
- `bolt.go` and `turret.go` define enemy fire: Turrets are immovable and fire Bolts (1 Health,
  shared Invulnerability, pass through enemies). Bolts keep flying, but cannot hit, while Dying and
  Game Over, and retry or session end clears them. Separation also keeps Turrets from overlapping
  each other; they are otherwise immovable. The scene's `RNG` makes first-volley offsets
  and Turret spawn spots deterministic in tests.
- `projectile.go`, `enemy.go`, and `player.go` define the combat entities. `lunge.go` holds the shared
  lunge rules (any player contact, even during Invulnerability, or the no-overlap floor ends a lunge; soft avoidance is skipped; a Barrel roll overlap is not contact) and `dart.go`
  the Dart's wind-up/charge/brake cycle. `sentinel.go` holds the Sentinel's orbit, telegraph and dash
  (direction and first-dash offset come from the scene `RNG` at spawn).
- Player controls: arrows/WASD steer, thrust and brake, Space fires the Blaster, X fires Rockets, Z
  starts a Barrel roll. A roll lasts 45 ticks with a 180-tick cooldown that starts with the roll.
  While rolling, Contact and Bolt resolution skip the player (no damage, no knockback, lunges are not
  ended, Bolts stay live); steering, thrust and firing still work and the ship does not move. Rolling
  is not Invulnerability and does not touch its timer. `Player.Reset` clears roll and cooldown, and
  Dying never rolls. The HUD shows a `Z Roll` cooldown bar below the weapon panel.
- Player visuals: the Interceptor (`ship_draw.go`) is drawn rotated to the heading with a blue engine
  glow driven by speed and `thrustHeld`. Blaster shots are twin tracers and Rockets classic missiles
  (`shot_draw.go`); the Scene records projectile impact effects (about 18 ticks) on enemy hits and
  clears them on reset. Invulnerability no longer blinks: the ship shakes, sparks, smokes and pulses
  red (`ship_hit_draw.go`), and a roll overrides that look (`ship_roll_draw.go`). Dying breaks the
  ship into four spinning pieces with a fireball (`ship_death_draw.go`).
- `symptoms.go` holds the cosmetic Symptoms: not ready flickers the eye and sparks the antenna,
  restarts weld 0/1/2 patch plates (0, 1–4, 5+), and a restart increase during play starts a
  20-tick reboot blink. Terminating and Exploding visuals hide them; they never change behaviour.
  Pending practice respawns keep updated Symptoms and never blink.
- `scene.go` owns the enemy pool, projectiles, practice-mode respawns, bridge event handling, and
  outbound kill events.
- `game.go` implements `ebiten.Game` and the top-level state machine.
- `draw.go` renders the screens, world, and HUD.
- `fieldguide.go` and `fieldguide_draw.go` own the Field guide: seven paged screens (Probe, Turret,
  Sentinel, Dart, Tiers & Bolts, Symptoms, and an illustrative `kubectl`-style table of how Pods
  become enemies). The archetype demos run real `Scene` behaviour against a dummy player. The guide
  reads no cluster data.

When adding a cross-layer event, define it in `game/events.go`, route it in `bridge/`, and add the
game-side consumer. Do not add cluster or bridge imports to `game/`.

## Session Lifecycle

```
StateTitle → StateModeSelect → StatePlaying → StateDying → StateGameOver
                   ↕
             StateFieldGuide
```

Mode select lists Practice, Chaos and, set apart below them, the Field guide, which is not a mode.
Enter on the guide enters `StateFieldGuide`; Escape or Backspace returns to mode select with the
cursor on that entry. The guide starts no session and opens no watch, so the bridge stays idle. The
cursor is frozen while the chaos confirmation shows, so chaos keeps its second press. The title
screen has no hint for the guide.

A session begins when the player confirms a mode. It owns one watch and has a monotonically
increasing ID. Escape from play or game over ends the session, clears the scene, cancels the watch,
and returns to the title. A watch-start error ends the session and returns to mode select with the
error displayed.

Retry from game over resets only the player and retains the active session and scene. In practice
mode, an enemy killed by a projectile returns after a delay because its Pod remains. In chaos mode, a
killed enemy does not respawn. If its deletion request fails, the running Pod is temporarily
unrepresented until a later provider event adds it again.

## CLI Flags

| Flag           | Default   | Description                                                                                                                                                                                                                                                     |
| -------------- | --------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `--mock`       | `false`   | Use the in-memory variety seed (nine eligible Pods plus an unmanaged one, with simulated replacements, not-ready replacements for ~3 s, and a ticker that flips `agent-y`'s readiness every ~10 s and adds a `web-c` restart every ~20 s) instead of a cluster. |
| `--namespace`  | `default` | Namespace watched for the session (also the namespace of the `--mock` seed).                                                                                                                                                                                    |
| `--kubeconfig` | `""`      | Explicit kubeconfig path; empty uses standard clientcmd loading rules.                                                                                                                                                                                          |
| `--context`    | `""`      | Kubernetes context override; empty uses the loaded config's current context.                                                                                                                                                                                    |

## Conventions

1. Use conventional commits.
2. `game/` never imports `cluster/` or `bridge/`.
3. Preserve typed channel boundaries between architecture layers.
4. Keep chaos mode as an explicit in-game, double-confirmed choice.
5. Surface watch-start errors in the game; log deletion failures. Startup failures in `main.go` are
   fatal and Ebitengine run failures are fatal.
6. Game tests construct game values and channels directly. Use `MockProvider` for cluster and bridge
   behavior tests.
