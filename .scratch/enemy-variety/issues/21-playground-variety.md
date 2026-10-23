# 21: Playground variety

**What to build:** The chaos playground provides every archetype and tier. See decision [Choosing mock and playground variety](10-choosing-mock-and-playground-variety.md).

**Blocked by:** 13

**Status:** resolved

Spec: [spec.md](../spec.md)

- [x] Deployment requests: none, 10m/16Mi, 50m/64Mi on two of them, and 250m/256Mi.
- [x] Add a StatefulSet with 2 replicas at 50m/64Mi, a DaemonSet at 10m/16Mi, and a Job running `pause` with `parallelism: 2`, `completions: 2`, `backoffLimit: 100` and 250m/256Mi.
- [x] Everything carries the existing playground labels. No RBAC changes.
- [x] The playground docs and README, if any, describe the expected mix.
- [x] The manifests validate (`kubectl apply --dry-run=client`, or kubeconform if available).

## Completion

- **Completed:** 2026-10-01
- **Commit:** `4eeb73f` (merged into `feat/enemy-variety-impl`)
- **Notes:** Validated with kubeconform -strict. The StatefulSet names a serviceName without creating a Service; the Job uses restartPolicy OnFailure.
