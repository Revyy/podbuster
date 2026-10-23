# Podbuster

Podbuster is a desktop space shooter in which every enemy is a Pod in a Kubernetes namespace. In
practice mode, shooting an enemy is harmless and the enemy comes back. In chaos mode, a kill sends a
real delete request for that Pod.

Podbuster is built with Go and [Ebitengine](https://ebitengine.org/).

## How Pods become enemies

Podbuster watches one namespace and spawns one enemy for each eligible Pod. A Pod is eligible when
a controller owns it and it has not finished (its phase is not `Succeeded` or `Failed`). Podbuster
skips Pods that no controller owns, because a deletion would remove them for good.

The controller kind picks the enemy's archetype:

| Controller    | Archetype | Behaviour                                |
| ------------- | --------- | ---------------------------------------- |
| StatefulSet   | Turret    | Stays in place and fires at you.         |
| DaemonSet     | Sentinel  | Circles you at a distance and dashes in. |
| Job           | Dart      | Winds up and charges in a straight line. |
| Anything else | Probe     | Flies toward you.                        |

Each enemy also gets a tier (Light, Standard, or Heavy). The tier compares the Pod's CPU and
memory requests with the median of the other Pods in the session. A Pod that is not ready, or that
has restarted, shows visible damage. That damage is cosmetic and does not change how the enemy
fights.

The in-game **Field guide** on the mode-select screen shows each archetype in action.

## Requirements

- Go 1.26.3 or later.
- On Linux, the X11 and OpenGL headers that Ebitengine needs. On Debian or Ubuntu:

	```bash
	sudo apt-get install libgl1-mesa-dev libxcursor-dev libxi-dev libxinerama-dev libxrandr-dev libxxf86vm-dev
	```

- For a real cluster, a kubeconfig whose user can list, watch, and delete Pods in the namespace.

## Play without a cluster

```bash
go run . --mock
```

The mock seeds nine eligible Pods that cover every archetype and tier, plus one Pod with no
controller that never spawns. Killed Pods come back as replacements, and the mock changes some
Pods' readiness and restart counts while you play.

## Play against a cluster

```bash
go run . --namespace=my-namespace
```

Podbuster uses your current kubeconfig context unless you pass `--context`. The game reads no
cluster data until you pick a mode.

### Try chaos mode safely

Chaos mode deletes real Pods. To try it without risk, use the local kind playground. It needs
`kind`, `kubectl`, and Docker or Podman.

```bash
make run-kind-chaos
```

This creates a kind cluster named `podbuster-chaos` with a namespace of throwaway workloads, then
starts the game against it. Run `make kind-reset` to restore the workloads and `make kind-down` to
delete the cluster. See [Kind chaos playground](docs/kind-chaos-playground.md) for details.

Chaos mode is not a command-line flag. Select it on the mode-select screen and press **Enter** a
second time to confirm. [ADR 0001](docs/adr/0001-chaos-mode-is-an-in-game-confirmed-choice.md)
explains why.

## Controls

| Key            | Action                                                  |
| -------------- | ------------------------------------------------------- |
| Arrows or WASD | Steer, thrust, and brake.                               |
| Space          | Fire the Blaster.                                       |
| X              | Fire Rockets.                                           |
| Z              | Barrel roll. Contact and Bolts miss you while you roll. |
| R              | Retry from the game-over screen.                        |
| Esc            | End the session and go back to the title.               |

## Command-line flags

| Flag           | Default   | Description                                                              |
| -------------- | --------- | ------------------------------------------------------------------------ |
| `--mock`       | `false`   | Use the in-memory mock Pods instead of a cluster.                        |
| `--namespace`  | `default` | The namespace to watch. With `--mock`, the namespace of the mock Pods.   |
| `--kubeconfig` | `""`      | The kubeconfig path. Empty uses `$KUBECONFIG` or `~/.kube/config`.       |
| `--context`    | `""`      | The kubeconfig context. Empty uses the current context.                  |

## Development

```bash
make test    # Build and run the tests with the race detector.
make lint    # Check go.mod, formatting, and golangci-lint.
make check   # Run lint and test. CI runs the same targets.
make fmt     # Apply formatting and lint autofixes.
```

`make lint` installs the pinned golangci-lint into `.bin/` the first time it runs.

The code has three layers. `cluster/` reads Pods from a cluster or the mock, `bridge/` turns Pod
events into game events, and `game/` runs the game. `game/` never imports the other two.
[AGENTS.md](AGENTS.md) describes the architecture and conventions, and [CONTEXT.md](CONTEXT.md)
defines the project's terms.

## License

Podbuster is released under the [MIT License](LICENSE).
