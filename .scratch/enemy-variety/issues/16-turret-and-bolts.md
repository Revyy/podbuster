# 16: Turret and Bolts

**What to build:** Turrets are stationary and immovable and fire Bolts with the decided per-tier patterns. See decisions [Specifying enemy fire](05-specifying-enemy-fire.md) and [Escalating archetypes by tier](06-escalating-archetypes-by-tier.md).

**Blocked by:** 15

**Status:** resolved

Spec: [spec.md](../spec.md)

- [x] Turrets never move. Contact pushes only the player, and Separation makes other enemies give way.
- [x] Light fires a fixed 4-way ring every 150 ticks. Standard fires a 6-way ring every 120 ticks, rotating 30° per volley. Heavy fires the Standard ring plus an aimed Bolt 60 ticks after each ring.
- [x] The ring telegraph lasts 30 ticks and the aimed Bolt's telegraph 15. The first volley comes after one interval plus a 0–60 tick offset from an injectable RNG on the scene.
- [x] Bolts move at 3 px/tick (later tuned to 2.4 in ticket 22), about 6 px across with a slightly smaller hitbox, in the firing archetype's colour. They expire at the screen edge.
- [x] A Bolt hit costs 1 Health and applies about half the contact knockback. It shares the Invulnerability window with contact. During Invulnerability a Bolt passes through and stays live.
- [x] Bolts pass through enemies and can't be shot down. A hit shows the ship flash and a small burst.
- [x] Terminating Turrets stop firing, and Bolts already in flight persist. There are no new volleys during Dying or Game Over, and in-flight Bolts can't hit then. Retry, reset and session end clear Bolts.
- [x] A Turret reverting from Terminating restarts its cycle. A practice respawn starts fresh.
- [x] `CONTEXT.md` terms are used throughout. Scene tests cover the patterns, rotation, aimed Bolt, collision and Invulnerability, pass-through, persistence and immovability. `go test ./...` is green.

## Completion

- **Completed:** 2026-10-01
- **Commit:** `5246dc6` (merged into `feat/enemy-variety-impl`)
- **Notes:** Turrets spawn on-screen, at least 60 px from the edges and 120 px from the screen centre, using the scene RNG; the spec left spawn position open. Cycles freeze during Dying. Checking it by hand is deferred to ticket 22.
