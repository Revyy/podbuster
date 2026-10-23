# Deciding live cosmetic overlays

Type: grilling
Status: resolved
Blocked by: 

## Question

Should a Pod's live signals change how its enemy looks after spawn, and if so, how? The live signals are readiness (Ready condition) and container restart count. Archetype and Tier stay fixed at spawn. Decide:
- which signals are shown, if any: not-ready, restart count, or a restart that happens during play
- how each one looks on the Seeker family, without clashing with the damage, contact, Terminating and Exploding visuals in [Prototyping archetype silhouettes](07-prototyping-archetype-silhouettes.md)
- whether an overlay is a lasting state (such as not-ready) or a brief flash (such as a restart that just happened)
- whether an overlay has any gameplay effect (default: none, it's cosmetic only)

Research input: [Surveying observable Pod traits](01-surveying-observable-pod-traits.md).

## Answer

- **Term**: these signals are **Symptoms** (added to `CONTEXT.md`). A Symptom is purely cosmetic, never changes stats or behaviour, and shows in both modes.
- **Not-ready Symptom**: shown while the Pod's `Ready` condition isn't True. While it lasts, the eye flickers irregularly (dims and stutters) and the antenna sparks now and then. The hull keeps full colour.
- **Restart Symptoms**: restarts are the sum of `restartCount` over the regular containers plus sidecar (restartable init) containers. Ordinary init containers are excluded.
  - **Lasting mark**: welded patch plates on the hull: none, one or two for 0, 1–4 or 5+ restarts. They are patches, not cracks, so they don't read as lost Hits.
  - **Reboot blink**: about 20 ticks, shown when the count rises during play. The eye shuts, the droid jolts, and the eye reopens.
- **Lifecycle**:
  - Terminating and Exploding visuals fully replace Symptoms.
  - An enemy that reverts after a failed deletion shows its current Symptoms again.
  - A freshly spawned not-ready enemy flickers. This is intended, because it shows a replacement Pod booting.
  - A practice respawn shows the Pod's current Symptoms with no blink for restarts that happened while it was away.
  - Restarts present at spawn show only as patches.
- **Validation**: no separate prototype. The [Designing the enemy info screen](11-designing-the-enemy-info-screen.md) prototype previews the Symptoms and tunes their look.
