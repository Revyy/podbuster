# Plan the implementation handoff

Type: grilling
Status: resolved
Blocked by: 01, 02, 03, 04, 05

## Question

Given the settled behavior and ownership decisions, what implementation ticket boundaries,
dependency order, and acceptance checks will let another agent build the weapons system without
reopening design questions?

## Answer

Five implementation tickets, created as `ready-for-agent` child issues:

1. [Refactor weapon slots and shared projectile model](07-refactor-weapon-slots-and-shared-projectile.md) —
   pure refactor: `Weapon` interface, Player-owned slots, and a shared cost-bearing projectile
   type, with no new player-visible behavior. Blocked by nothing; the pure-refactor scope was kept
   separate from the rocket weapon so a half-merged cost model is never player-visible.
2. [Implement the rocket volley secondary weapon](08-implement-rocket-volley-weapon.md) — the X
   input, three-rocket pattern, cooldown, and visuals. Blocked by 07.
3. [Build the weapon status HUD panel](09-build-weapon-status-hud-panel.md) — the bottom-left
   loadout panel with the cooldown fill bar. Blocked by 07 and 08 (needs real weapon names and
   cooldown state to read).
4. [Wire weapon lifecycle resets](10-wire-weapon-lifecycle-resets.md) — cooldown reset in
   `Player.Reset`, confirming the death-animation freeze and lifecycle isolation. Blocked by 07
   and 08.
5. [Final integration pass](11-final-integration-pass.md) — sweep, full test run, manual playtest,
   doc cross-references. Blocked by 09 and 10.

Tickets 3 and 4 (HUD panel, lifecycle resets) are parallelizable once 1 and 2 land: they touch
disjoint files (`draw.go`/new HUD code vs. `Player.Reset`) and neither result depends on the
other, mirroring the map's own blocking graph where 04 and 05 shared only ticket 01 as a
prerequisite. Tickets 1 and 2 stay strictly linear since the rocket weapon is built on the
refactored `Weapon`/projectile model.

Each implementation ticket carries its own acceptance criteria (test coverage, behavior checks,
and for the HUD ticket, a pointer to the archived prototype on `prototype/weapon-hud-status` for
reference only, not to be merged).
