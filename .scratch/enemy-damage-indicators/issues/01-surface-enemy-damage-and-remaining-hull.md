# 01: Surface enemy damage and remaining hull

**What to build:** Each Pod enemy visibly reacts when struck by a bullet and continuously communicates its remaining hits, so a player can distinguish a freshly spawned enemy from one close to destruction. The display preserves the existing rule that contact only causes knockback and does not damage Pods.

**Blocked by:** None (can start immediately).

**Status:** ready-for-agent

- [x] A bullet hit gives the struck enemy a visible damage reaction distinct from the existing contact reaction.
- [x] Every active enemy visibly shows how many bullet hits remain before it is destroyed.
- [x] Enemy damage feedback resets correctly when a practice-mode enemy respawns and does not let contact damage change the displayed hit count.
