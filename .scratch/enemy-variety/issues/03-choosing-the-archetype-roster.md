# Choosing the archetype roster

Type: grilling
Status: resolved
Blocked by: 01

## Question

Which archetypes exist, which Pod traits select each (primarily owner kind: Deployment, StatefulSet, DaemonSet, Job/CronJob, bare Pod), and what does each do differently in look and behaviour? Example from charting: a stationary StatefulSet turret that fires in a fixed pattern. Also decide whether "consequence of deletion" should shape the roster.

## Answer

- **Archetype source**: the Pod's direct controller owner kind (`ownerReferences[controller=true]`). No ReplicaSet→Deployment or Job→CronJob resolution, so no extra watches or RBAC. ReplicaSet counts as Deployment-like; Job covers CronJob.
- **Roster.** All four share the Seeker construction (orb, eye, tentacles), each with its own silhouette and hue. Tier varies size and detail within an archetype.
  - **Probe**: ReplicaSet, plus any unrecognised controller kind as the fallback. A homing Seeker, as in the prototype.
  - **Turret**: StatefulSet. Stationary; fires a fixed radial pattern. Squat plated hull, tentacles planted like legs, eye on a barrel.
  - **Sentinel**: DaemonSet. Orbits the player at range and periodically dashes in. A ring around the orb.
  - **Dart**: Job (covers CronJob). Fast, low Hits; charges in straight lines and overshoots. Slim and swept-back, with long trailing tentacles.
- **Unmanaged Pods** (no controller owner) are **not eligible** in either mode, because deleting them in chaos mode is permanent and that doesn't fit chaos-monkey play. So there is no Stray archetype. The exclusion must sit behind a single eligibility rule, so a future configuration option could include them.
- **Managed** means any controller owner, including Job Pods, even though deleting one may count against the Job's `backoffLimit` (the info screen notes this) and unknown operators aren't guaranteed to recreate. Excluding these may be revisited later.
- **Deletion consequence** does not shape difficulty.
- **StatefulSet ordinals** get no special treatment; difficulty belongs to Tier.
- **Spec scope**: all four archetypes. Implementation order is Probe → Turret → Dart → Sentinel.
- **Glossary**: Pod's definition now requires being managed and stays a Pod while terminating; added Managed Pod, Turret, Sentinel and Dart.
