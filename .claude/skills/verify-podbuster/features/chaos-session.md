# Chaos session

Chaos is the double-confirmed mode in which killing an enemy requests deletion of its Pod. The
HUD shows `!! CHAOS MODE !!` and counts `Pods deleted`. Killed enemies don't respawn as the same
Pod, but a controller replacement Pod arrives as a new enemy.

## Sub-features

- `chaos-start` needs two Enters on Chaos. The HUD shows the `!! CHAOS MODE !!` banner and `Pods deleted: 0`.
- `chaos-kill-deletes` makes a kill raise `Pods deleted` and leaves the Pod terminating, then gone, in the cluster.
- `chaos-replacement` spawns the controller's replacement Pod as a new enemy.
- `chaos-delete-failure` logs a failed deletion in `game.log` as `bridge: failed to delete pod …` and reverts the enemy.

## How to get to it (user POV)

- Title → Enter → down → Enter (warning) → Enter (confirm).
- `R` on a chaos game-over screen.

## Driving it with podctl

Preconditions:

- Docker (or Podman with `KIND_EXPERIMENTAL_PROVIDER=podman`) is running.
- `make kind-up` succeeded. It creates or reuses the `podbuster-chaos` cluster and namespace.
- The run was launched with `$P launch <id> kind`, which uses `--context kind-podbuster-chaos --namespace podbuster-chaos`.

- **Baseline pods.** Run `kubectl --context kind-podbuster-chaos -n podbuster-chaos get pods -o wide > .verify/runs/<id>/pods-before.txt`. Ten Running Pods are listed (see docs/kind-chaos-playground.md).
- **Arm and confirm.** Run `$P key <id> 80 enter; sleep 0.3; $P key <id> 80 down; sleep 0.3; $P key <id> 80 enter; sleep 0.3; $P shot <id> chaos-warning; $P key <id> 80 enter; sleep 1; $P shot <id> chaos-start`. The first frame shows the DELETE warning for `podbuster-chaos`. The second shows `!! CHAOS MODE !!`, `Pods deleted: 0` and `Enemies: 10`.
- **Kill.** Run `$P spray <id> 100; $P shot <id> after-spray`. The frame shows `Pods deleted:` ≥ 1.
- **Cluster side effect.** Run `kubectl --context kind-podbuster-chaos -n podbuster-chaos get pods -o wide > .verify/runs/<id>/pods-after.txt` and diff it against `pods-before.txt`. At least one Pod name from before is gone or Terminating, and a replacement with a new name appears for Deployment/StatefulSet/DaemonSet/Job owners.
- **No failures.** `grep bridge: .verify/runs/<id>/game.log` prints nothing.
- **Reset.** Run `make kind-reset` so the next run starts from the documented set of targets.

## Gotchas

- Never launch chaos against any context other than the kind playground. `podctl` only offers `mock` and `kind` for this reason.
- `kind` mode needs the playground before launch. Launch fails with `kind playground not reachable` otherwise.
- `Pods deleted` counts kill requests. Only `kubectl` proves the deletion happened.
- Not yet run through this skill: the first chaos verification should record any extra wait (for example, the Job Pods' replacement time) here.
