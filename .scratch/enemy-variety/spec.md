# Spec: Enemy variety

Status: ready-for-agent

Charted by the wayfinder map (`map.md`). The decision tickets 01–11 hold each decision in full, and this spec summarises them. Where the two disagree, the decision ticket wins.

## Problem Statement

Every enemy looks and behaves the same, whatever its Pod is. A player shooting at a namespace can't tell a StatefulSet replica from a batch Job, or a heavyweight Pod from a tiny one. The game also says nothing about Pod health. The cluster's structure is invisible, and play is monotonous.

## Solution

Each enemy's **Archetype** comes from its Pod's controller kind, and its **Tier** from its resource requests relative to the namespace. Both are fixed at spawn.
- **Probe**: ReplicaSet Pods and any unrecognised controller.
- **Turret**: StatefulSet Pods.
- **Sentinel**: DaemonSet Pods.
- **Dart**: Job Pods.

Every archetype is a member of the vector-drawn "Seeker" droid family, with its own silhouette, colour and behaviour. Turrets fire **Bolts**. Tiers (Light, Standard, Heavy) scale the enemy's stats and visibly change its size and armour. Live Pod health shows as cosmetic **Symptoms**. The **Field guide**, opened from mode select, teaches all of this. Unmanaged Pods are no longer enemies. Mock mode and the chaos playground provide a representative mix.

## User Stories

1. As a player, I want enemies to look different by the kind of workload behind them, so that I can read the namespace at a glance.
2. As a player, I want Deployment (ReplicaSet) Pods to be Probes that home in on me, so that the baseline enemy stays familiar.
3. As a player, I want StatefulSet Pods to be stationary Turrets that fire Bolt rings, so that stateful workloads feel entrenched.
4. As a player, I want DaemonSet Pods to be Sentinels that orbit me and dash in, so that node-level agents feel watchful.
5. As a player, I want Job Pods to be fragile Darts that charge and overshoot, so that batch work feels fast and short-lived.
6. As a player, I want Pods with unrecognised controllers to appear as Probes, so that no managed Pod is missing.
7. As a player in chaos mode, I want Pods with no controller to never appear, so that I can't permanently delete something nobody will recreate.
8. As a player, I want heavier-requesting Pods to be bigger, armoured and tougher, so that resource weight is visible and matters.
9. As a player, I want Pods without requests to be Light, so that unconfigured workloads read as flimsy.
10. As a player, I want tiers judged against the other Pods in the namespace, so that every namespace gets a spread of difficulty.
11. As a player, I want an enemy's Archetype and Tier never to change after spawn, so that what I've learned about it stays true.
12. As a player, I want Light enemies to be smaller, thinner and quicker, so that I can tell them from Heavy ones at a glance.
13. As a player, I want Heavy enemies to carry armour bands and an extra tentacle, so that the threat is legible.
14. As a player, I want Heavy contact to cost 2 Health, so that heavyweights are worth avoiding.
15. As a player, I want Darts to die in fewer hits than other archetypes, so that their speed is balanced by fragility.
16. As a player, I want a Turret's eye to glow before each volley, so that I can dodge fairly.
17. As a player, I want Light Turrets to fire a fixed 4-way ring, Standard a rotating 6-way ring, and Heavy to add an aimed Bolt, so that harder Turrets demand more.
18. As a player, I want Bolts to be slower than my ship, so that dodging is always possible.
19. As a player, I want a Bolt hit to cost 1 Health and share contact's Invulnerability, so that damage stays predictable.
20. As a player, I want Bolts to pass through enemies and be unshootable, so that the rules stay simple.
21. As a player, I want Bolts already in flight to persist after their Turret dies, so that killing a Turret doesn't erase danger unrealistically.
22. As a player, I want Turrets to be immovable, so that "stationary" means stationary.
23. As a player, I want Sentinels to telegraph their dash and lock my position, so that I can sidestep.
24. As a player, I want heavier Sentinels to orbit tighter and dash more often.
25. As a player, I want Darts to wind up visibly before charging, so that I can read their attack.
26. As a player, I want heavier Darts to charge faster and further with a shorter wind-up.
27. As a player, I want every lunge to stay at or below my top speed, so that I can always escape.
28. As a player, I want contact to end a lunge, so that I'm never pinned.
29. As a player, I want a not-ready Pod's enemy to flicker its eye and spark, so that I can see unhealthy workloads.
30. As a player, I want restart-prone Pods to show patch plates, so that crash history is visible.
31. As a player, I want a reboot blink when a Pod restarts during play, so that live events register.
32. As a player, I want Symptoms never to change behaviour, so that they are information, not mechanics.
33. As a player, I want a Field guide in mode select, so that I can learn the archetypes, tiers, Bolts and Symptoms before playing.
34. As a player, I want the Field guide to show each archetype in action, so that I learn by seeing.
35. As a player, I want the Field guide to map example `kubectl` rows to enemies, so that I understand where enemies come from.
36. As a player, I want the Dart page to warn that deleting a Job Pod may count against its retry limit.
37. As a player in practice mode, I want a respawned enemy to keep its Archetype and Tier and show current Symptoms.
38. As a player in chaos mode, I want a replacement Pod to get a fresh tier and visibly boot (not ready).
39. As a developer, I want `--mock` to seed every archetype, tier and Symptom, so that I can see the full variety without a cluster.
40. As a developer, I want the mock to simulate Symptom changes and replacement Pods, so that live behaviour is visible offline.
41. As a developer, I want the chaos playground to include a StatefulSet, a DaemonSet, a Job and varied requests, so that chaos mode can show every archetype and tier.
42. As a developer, I want initial tiers computed after the watch has synced, so that tiers don't depend on arrival order.
43. As a developer, I want one eligibility rule in one place, so that including unmanaged Pods can later become a config option.
44. As a developer, I want `game/` to stay free of Kubernetes concepts, so that the layer boundary holds.

## Implementation Decisions

- **Layering** ([Shaping the trait payload](issues/09-shaping-the-trait-payload.md)):
  - Providers report raw traits in `PodInfo`: controller owner kind, CPU and memory requests, ready, and restarts.
  - The bridge owns the single eligibility rule (managed and non-terminal), derives Archetype and Tier, and keeps the per-session map that the median is computed from.
  - Game events carry Archetype, Tier and Symptoms. Their enums live in the game's events file.
  - A new Symptoms-changed event is emitted only when ready or restarts change.
  - The provider watch gains a synced marker. The bridge holds initial Pods until it arrives. The real provider returns its channel immediately and sends the marker once the handler has synced.
- **Archetype mapping** ([Choosing the archetype roster](issues/03-choosing-the-archetype-roster.md)): the direct controller owner kind only. ReplicaSet and unknown kinds become Probe, StatefulSet becomes Turret, DaemonSet becomes Sentinel, and Job becomes Dart.
- **Tier banding** ([Defining tier banding](issues/04-defining-tier-banding.md)):
  - The score is the larger of the CPU and memory request ratios to the namespace median. Below ½ is Light, ½ to 2 is Standard, and above 2 is Heavy.
  - A Pod with no requests is Light and is excluded from the median.
  - A request change updates the median but never re-tiers an existing enemy.
- **Shared tier stats and base stats per archetype** ([Escalating archetypes by tier](issues/06-escalating-archetypes-by-tier.md)):
  - Hits are 2/3/5 (Dart 1/2/3).
  - Size is 0.85×/1×/1.3×, and the hitbox scales with it.
  - Cruising speed is 1.1×/1×/0.85×. It doesn't apply to dashes or charges.
  - Contact damage is 1/1/2.
  - Separation must account for each pair's sizes.
- **Bolts and Turret** ([Specifying enemy fire](issues/05-specifying-enemy-fire.md) and ticket 06): see those tickets for the speeds, timings, telegraphs, ring patterns, Invulnerability sharing, knockback and lifecycle.
- **Sentinel and Dart behaviour**: ticket 06, including the slower Sentinel speeds from playtesting, the lunge rules, and the timer rules for respawn and revert.
- **Visuals** ([Prototyping the Probe droid](issues/02-prototyping-the-probe-droid.md) and [Prototyping archetype silhouettes](issues/07-prototyping-archetype-silhouettes.md)):
  - Port the vector drawings from branches `prototype/probe-droid` and `prototype/archetype-silhouettes`, including the tier cues and state animations.
  - Droids are drawn upright and may overhang their hitbox.
- **Symptoms** ([Deciding live cosmetic overlays](issues/08-deciding-live-cosmetic-overlays.md)):
  - Not ready makes the eye flicker and the antenna spark.
  - Restarts show as patch plates: none, one or two for 0, 1–4 or 5+.
  - A restart during play triggers a 20-tick reboot blink.
  - Terminating and Exploding visuals hide Symptoms.
  - The game does the banding and detects restart increases.
- **Mock and playground** ([Choosing mock and playground variety](issues/10-choosing-mock-and-playground-variety.md)): the 9+1 seed table, the new mock constructor and Symptoms setter, the `--mock`-only ticker and replacement Pods, and the playground manifests.
- **Field guide** ([Designing the enemy info screen](issues/11-designing-the-enemy-info-screen.md)): a third entry on mode select and a 7-page codex. The prototype is on branch `prototype/enemy-info-screen`.
- **Randomness**: Turret and Sentinel first-cycle offsets come from an injectable RNG on the scene, so tests stay deterministic.

## Testing Decisions

- Test external behaviour at existing seams only. Don't test drawing code.
- **Bridge seam** (MockProvider → Bridge → game events):
  - eligibility, including unmanaged Pods and a Pod becoming ineligible;
  - spawns held until the synced marker;
  - Archetype mapping;
  - tier banding cases (uniform namespace, no requests, a lone Pod, Heavy and Light ratios);
  - the median after deletes;
  - Symptoms-changed emitted only on real changes;
  - stale-session protections unchanged.

  Prior art: the existing bridge tests.
- **Scene seam** (construct the scene, feed game events, step ticks with fixed controls and RNG):
  - tier stats;
  - Turret volleys, rotation and aimed Bolt;
  - Bolt collision, shared Invulnerability, pass-through and persistence;
  - immovable Turrets;
  - Dart and Sentinel cycles and lunge termination;
  - respawn keeping Archetype and Tier.

  Prior art: the existing scene and enemy tests.
- **KubeClient seam** (fake clientset): traits in `PodInfo`, request summing, readiness and restarts, the synced marker, and no blocking with more than 100 Pods.
- Every ticket ends with `go test ./...` green.

## Out of Scope

- Including unmanaged Pods, or a "Stray" archetype.
- Resolving ReplicaSet→Deployment or Job→CronJob owners.
- Hover or crosshair Pod labels.
- Symptoms with gameplay effects.
- Re-tiering after spawn.

## Further Notes

- All numbers are starting values to tune.
- Update the `cluster/` and `bridge/` sections of AGENTS.md to describe the new filtering and events.

## Implementation

- **Completed:** 2026-10-01
- **Branch:** `feat/enemy-variety-impl`. There is no remote, so no PR.
- **Tickets resolved:** 12–22. See each ticket's Completion section for its commits and judgement calls.
- **Changes from this spec:**
  - Bolts fly at 2.4 px/tick instead of 3, after the manual check.
  - Contact ends a lunge even while the player is invulnerable.
  - Turrets push each other apart to avoid overlap.
