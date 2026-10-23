# Chart an extensible two-slot weapons system

Label: wayfinder:map

## Destination

An implementation-ready design and ordered ticket plan for one main weapon and one secondary
weapon. The first secondary weapon fires a three-rocket volley, while the design leaves a clean
replacement point for future weapons without specifying a power-up system.

## Notes

- Consult `domain-modeling` in every session and keep `CONTEXT.md` current as terms settle.
- Use `grilling` for decision tickets and `prototype` for tickets about game feel or presentation.
- The main weapon keeps the current Space input and basic one-Hit projectile behavior.
- X fires one volley of three rockets on parallel tracks.
- Each rocket stops at its first enemy and costs that enemy two Hits. Rockets do not home, splash,
  pierce, or knock enemies back.
- Rocket volleys are unlimited and governed by a cooldown.
- The first iteration has explicit main and secondary weapon slots and a shared projectile impact
  model.
- The HUD names both equipped weapons and shows whether the secondary weapon is ready.
- The final design must cover behavior, domain terms, game-package ownership, rendering, HUD,
  acceptance criteria, and ordered implementation tickets.

## Decisions so far

- [Define weapon slot ownership](issues/01-define-weapon-slot-ownership.md): Player owns the two
  equipped slots and routes input; weapons own firing state and create projectiles; Scene simulates
  projectiles and resolves their impacts.
- [Tune the rocket volley](issues/02-tune-the-rocket-volley.md): X fires a three-rocket burst with a
  center shot and two ±8° side lanes, a 9.0 px/tick rocket speed, 10×10 projectile size, and a
  180-tick cooldown, keeping the main weapon as the default while the secondary adds burst damage.
- [Define projectile impact rules](issues/03-define-projectile-impact-rules.md): projectiles carry a
  weapon-specific hit cost, each collision resolves once and only through the shared scene impact
  loop, and contact never deletes a Pod; rockets cost 2 hits and overkill clamps cleanly at zero.
- [Design weapon status in the HUD](issues/04-design-weapon-status-in-the-hud.md): a bottom-left
  loadout panel names both slots with their inputs and shows the secondary's cooldown as a
  color-shifting fill bar, staying clear of every existing HUD element.
- [Define weapon lifecycle rules](issues/05-define-weapon-lifecycle-rules.md): equip is fixed for
  the Game instance's life; cooldown resets to ready on session start and retry, freezes during
  the death animation, and stays untouched by practice-mode respawns or chaos-mode deletion.
- [Plan the implementation handoff](issues/06-plan-the-implementation-handoff.md): five ordered
  implementation tickets (refactor → rocket weapon → HUD panel + lifecycle resets in parallel →
  final integration), each `ready-for-agent` with its own acceptance criteria.

## Not yet specified

- None at the current frontier.

## Out of scope

- Power-up spawning, collection, inventories, weapon acquisition, swapping, and upgrades.
- Main weapons other than the current basic weapon.
- Secondary weapons other than the first rocket volley.
- Changes to Pod watching, deletion, session identity, or cross-layer channels.
