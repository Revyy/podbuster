# Prototyping archetype silhouettes

Type: prototype
Status: resolved
Blocked by:

## Question

What do the Turret, Sentinel, and Dart look and move like as members of the Seeker family? And how do the Light and Heavy tier cues read in play?

Reference designs:
- Turret: squat plated hull, tentacles planted like legs, eye on a barrel.
- Sentinel: a ring around the orb; orbits the player and dashes in.
- Dart: slim and swept-back, with long trailing tentacles; charges straight and overshoots.
- Tier cues: Light 0.85× with a thinner hull; Heavy 1.3× with armour bands and an extra tentacle.

Build on the Seeker prototype on branch `prototype/probe-droid` (`drawProbeSeeker`). Use rough movement for the Sentinel orbit/dash and the Dart charge so the silhouettes can be judged in motion. For Turret fire, use the rules resolved in "Specifying enemy fire": Bolts are about 6 px glowing pellets in the Turret's hue at 3 px/tick, fired as a 6-way ring every 120 ticks, with a 30-tick eye glow before each volley.

## Answer

The reference designs work as members of the Seeker family, and the tier cues read in play, not just in the gallery.

- **Probe** (red): the Seeker unchanged.
- **Turret** (amber): squat, with a wide dome on a plated skirt and legs planted outward. Its eye sits on a barrel that aims at the player. The eye glows brighter before each volley.
- **Sentinel** (green): an orb with a tilted ring of orbiting nodes that passes in front of and behind the orb. The ring flashes white before a dash. In motion, the decided speeds were too fast and were lowered by a third: a 0.8 px/tick orbit and a 2.7 px/tick dash. [Escalating archetypes by tier](06-escalating-archetypes-by-tier.md) is updated to match.
- **Dart** (violet): a slim body with a nose spike and a swept-back top fin, and three long, thin tentacles. It faces its direction of travel, and its eye glows during the wind-up.
- **Tier cues**:
  - Light: 0.85× scale, hull 0.8× as wide, thinner trim and tentacles.
  - Heavy: 1.3× scale, dark metal armour bands with rivets, and one extra tentacle (an extra leg on the Turret) and a thicker Sentinel ring.
- **States**: each archetype adapts the Probe's state language. The Turret's barrel droops when it is Terminating, and the Sentinel's ring and the Dart's fin split off with the hull halves when they explode.

Prototype (primary source): branch `prototype/archetype-silhouettes`, file `game/prototype_archetypes.go`, commits `2e2a59c` and `7110f32`.

Run it with `go run -tags prototype . --mock`, then start practice and press M. Keys:
- G: gallery.
- F1–F4: pick a gallery row.
- F5: cycle that row's state.
- Tab: cycle the archetype filter.
- T: reshuffle.

The prototype's motion is approximate. Ticket 06 holds the authoritative behaviour values.
