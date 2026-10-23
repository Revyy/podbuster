# Enemy variety

Label: wayfinder:map

## Destination

A spec, ready to hand off as implementation tickets, for Pod-driven enemy variety. It covers the baseline **Probe** look; the Pod traits that set each enemy's **Archetype** and **Tier**; how both map to visuals and difficulty; and an info screen that explains the enemies.

## Notes

- Domain: game design and Kubernetes. Use the `CONTEXT.md` terms: Archetype, Tier, and Probe.
- Skills: `grilling` + `domain-modeling` for decisions, and `prototype` for visuals.
- Plan, don't do. Building is handed off once the spec is clear.
- Standing decisions from charting:
  - Difficulty uses the existing levers (Hits, speed, contact damage, size) plus a small set of new behaviours. Example: a stationary StatefulSet turret that fires in a fixed pattern, with harder tiers aiming at the player or firing faster.
  - Traits split across axes, with owner kind driving Archetype and resources driving Tier. This is the leading hypothesis; tickets 03 and 04 confirm or revise it.
  - Tier is relative within the session's namespace and banded. Pods without requests fall into the lowest tier.
  - Archetype and Tier are fixed at spawn. Live traits such as restarts and readiness may only change cosmetic overlays.
  - The droid is drawn upright (side view) whatever its heading. It may overhang its 24×24 hitbox; the baseline hitbox is unchanged.
  - Traits must be readable from an enemy's look and are taught by an info screen selectable from the menu: each archetype, what it does, and what it originates from.
  - Mock mode must seed a representative mix of archetypes and tiers.
  - The baseline covers the static look, idle animation, and droid-specific Terminating, Exploding, damage, and contact visuals.
  - `game/` stays free of `cluster/`/`bridge/` imports, so traits must cross via `game/events.go`.

## Decisions so far

- [Surveying observable Pod traits](issues/01-surveying-observable-pod-traits.md): Use owner kind for Archetype and namespace-relative resource requests for Tier; keep QoS as fallback and readiness/restarts as cosmetic overlays.
- [Prototyping the Probe droid](issues/02-prototyping-the-probe-droid.md): The Probe is the vector-drawn "Seeker": a hovering orb with a player-tracking eye and trailing tentacles. Sprites are not needed.
- [Choosing the archetype roster](issues/03-choosing-the-archetype-roster.md): Direct owner kind picks the archetype: ReplicaSet/unknown → Probe, StatefulSet → Turret, DaemonSet → Sentinel, Job → Dart. Unmanaged Pods are excluded behind a single, future-configurable eligibility rule.
- [Defining tier banding](issues/04-defining-tier-banding.md): Each enemy gets one of three tiers, Light, Standard or Heavy, from the larger of its CPU/memory request ratios to the namespace median. A Pod without requests is Light. The tier is fixed at spawn, and it scales Hits, size, speed and contact damage.
- [Specifying enemy fire](issues/05-specifying-enemy-fire.md): Enemy shots are Bolts: 1 Health, shared Invulnerability, slow and dodgeable, not blockable or shootable, persisting after their source dies. The Standard Turret fires a rotating 6-way ring every 2 s after a glow telegraph and is immovable.
- [Escalating archetypes by tier](issues/06-escalating-archetypes-by-tier.md): Each archetype has base stats (the Dart has fewer Hits), and the tier speed lever scales cruising only. Turret: a static 4-ring at Light, a rotating 6-ring at Standard, and the ring plus an aimed Bolt at Heavy. Sentinel: tighter orbit and more frequent dashes as tier rises. Dart: faster charges, longer overshoot and shorter wind-up. Only the Turret fires.
- [Prototyping archetype silhouettes](issues/07-prototyping-archetype-silhouettes.md): The reference silhouettes and tier cues work in play: amber Turret, green ringed Sentinel, violet swept Dart. Playtesting cut Sentinel speeds by a third.
- [Deciding live cosmetic overlays](issues/08-deciding-live-cosmetic-overlays.md): Readiness and restarts show as cosmetic Symptoms: a flickering eye when not ready, patch plates for restart count, and a reboot blink when a restart happens during play. Terminating and Exploding visuals hide them.
- [Shaping the trait payload](issues/09-shaping-the-trait-payload.md): Providers report raw traits. The bridge owns the single eligibility rule, derives Archetype and Tier, and holds spawns until a new synced marker. Game events carry Archetype, Tier and Symptoms, with a new Symptoms-changed event.
- [Choosing mock and playground variety](issues/10-choosing-mock-and-playground-variety.md): `--mock` seeds a fixed 9-Pod mix covering every archetype, tier and Symptom, plus an unmanaged Pod that is never shown. It simulates Symptom changes and replacement Pods. The playground adds a StatefulSet, a DaemonSet, a Job and varied requests.
- [Designing the enemy info screen](issues/11-designing-the-enemy-info-screen.md): The Field guide is a 7-page codex reached from a mode-select entry: four archetype pages with live demos, Tiers & Bolts, Symptoms, and an example kubectl table that maps Pods to enemies.

## Handoff

Destination reached: [spec.md](spec.md), with implementation tickets 12–22 in `issues/`.

## Not yet specified

_(Nothing remains here: the four earlier entries are now the tickets "Deciding live cosmetic overlays", "Shaping the trait payload", "Choosing mock and playground variety" and "Designing the enemy info screen".)_

## Out of scope

- Including unmanaged Pods (no controller), or any "Stray" archetype: excluded because deleting them is permanent. They may return later as a configuration option, so the spec only keeps the eligibility rule in one place ([Choosing the archetype roster](issues/03-choosing-the-archetype-roster.md)).
- Resolving ReplicaSet→Deployment or Job→CronJob owners: not needed to separate archetypes ([Choosing the archetype roster](issues/03-choosing-the-archetype-roster.md)).
- Hover or crosshair labels identifying an enemy's Pod. The info screen was chosen instead.
- Implementation. The map ends at a handoff-ready spec.
