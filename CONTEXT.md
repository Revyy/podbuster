# Podbuster

A desktop space shooter in which every enemy on screen stands for a Pod running in a Kubernetes
cluster. Shooting an enemy can, in one mode, delete the real Pod. This glossary fixes the words
we use to keep the game world and the cluster world distinguishable.

## Language

### Cluster and game world

**Pod**:
A non-terminal, managed workload in the cluster. It remains a Pod while terminating, until it is
finally removed. Podbuster observes Pods but never creates them.
_Avoid_: container, workload, instance

**Managed Pod**:
A Pod with a controller that recreates it if deleted. Only managed Pods are represented as
enemies; a Pod no controller owns is left out of the game, because deleting it would be
permanent.
_Avoid_: owned pod, replicated pod

**Enemy**:
The on-screen adversary created for one eligible Pod reported by the active namespace watch.
_Avoid_: mob, ship, alien

**Enemy pool**:
The set of enemies currently on screen. It represents eligible Pods in the watched namespace,
except while a practice-mode enemy waits to respawn.
_Avoid_: enemy list, spawn list

**Namespace**:
The single Kubernetes namespace a session observes. Podbuster watches one at a time.

**Archetype**:
The kind of enemy a Pod becomes, fixed when the enemy spawns. It determines the enemy's
look and behaviour.
_Avoid_: variant, class, species, enemy type

**Probe**:
The baseline archetype: the plainest, easiest enemy. It is a hovering droid with one eye that
watches the player and tentacles that trail behind it.
_Avoid_: basic enemy, default enemy

**Turret**:
The archetype for StatefulSet Pods: a stationary enemy that fires at the player.
_Avoid_: tower, cannon

**Sentinel**:
The archetype for DaemonSet Pods: it circles the player at a distance and dashes in.
_Avoid_: guard, orbiter

**Dart**:
The archetype for Job Pods: fast, fragile, and charging in straight lines.
_Avoid_: kamikaze, rusher

**Tier**:
An enemy's difficulty band within its archetype, fixed at spawn and judged relative to the other
Pods in the namespace rather than by absolute thresholds. There are three tiers: Light, Standard,
and Heavy.
_Avoid_: level, rank, strength

**Symptom**:
A cosmetic sign on an enemy of its Pod's live health: not ready, or restarting. A Symptom can
appear or clear at any time after spawn, and never changes an enemy's stats or behaviour.
_Avoid_: status effect, debuff, overlay, condition

### Combat

**Hits**:
The number of projectile impacts an enemy can absorb before it is killed. Counted down per enemy.
_Avoid_: enemy health, enemy hitpoints

**Health**:
The player ship's damage pool, reduced by contact with enemies and by Bolt impacts. Reaching zero
ends the session's play.
_Avoid_: player hits, lives, hitpoints

**Kill**:
The in-game event of an enemy's hits reaching zero. A kill is caused only by projectiles; ramming an
enemy never kills it. A kill moves its enemy into Terminating rather than removing it immediately.
_Avoid_: destroy, frag

**Terminating**:
The enemy state entered the instant its Hits reach zero, or when the namespace watch reports its
Pod already terminating in the cluster for any other reason. A Terminating enemy is frozen and
harmless: it stops homing and firing, deals no contact damage, and is excluded from Separation.
Bolts it fired earlier stay in flight. It dims and
pulses until the Pod is confirmed gone, or reverts to active if a chaos-mode deletion request
fails before anything real starts. Applies in both modes, and to any observed Pod regardless of
who or what caused its termination.
_Avoid_: dying (reserved for the player), death, expiring

**Exploding**:
The enemy's brief, final, one-shot destruction animation, entered once its Pod is confirmed gone.
In practice mode, where no deletion is ever requested, Terminating and Exploding are both played
on a short fixed timer for visual consistency rather than driven by real cluster state.
_Avoid_: death animation (reserved for the player), destroyed

**Projectile**:
A shot created by an equipped weapon. Each projectile carries its weapon's hit cost and stops after
its first enemy impact.
_Avoid_: bullet

**Bolt**:
A shot fired by an enemy. It costs the player Health on impact and never harms enemies. It passes
through enemies and cannot be shot down; once fired, it no longer depends on the enemy that fired it.
_Avoid_: enemy projectile, enemy bullet

**Deletion**:
The requested cluster consequence of a kill: Podbuster asks Kubernetes to remove the real Pod.
Only chaos mode requests deletions, and a request can fail. A kill is not a deletion.
_Avoid_: pod kill, reap

**Contact**:
An overlap between the player ship and an enemy. It costs the player health and pushes both
parties apart; it never costs the enemy hits. Only the player can be party to contact.
_Avoid_: collision, ram

**Separation**:
The distance enemies keep from one another, so that every enemy stays a distinct target the player
can shoot and ram individually. Enemies crowd each other but never overlap; separation is not
contact and costs nobody hits or health.
_Avoid_: spacing, repulsion, collision, personal space

**Invulnerability**:
The window after contact or a Bolt impact during which the player cannot take further damage from
either.
_Avoid_: i-frames, cooldown, shield

**Barrel roll**:
The player's brief defensive ability, available again after a cooldown. While the ship rolls,
Bolts and enemies pass through it without contact or damage. It is not Invulnerability, which
only follows damage. The roll does not move the ship.
_Avoid_: dodge, dash, evade, i-frames

### Modes and lifecycle

**Session**:
One run of play, from mode confirmation until return to the title screen or a watch-start failure.
A session owns exactly one namespace watch; no cluster data is read outside a session.
_Avoid_: round, match, game

**Practice mode**:
The mode in which kills have no cluster consequence. Because the Pod survives its enemy's kill,
the enemy returns to the pool after a delay.
_Avoid_: safe mode, dry run, read-only mode

**Chaos mode**:
The mode in which a kill requests deletion of the real Pod. Chosen deliberately in-game and
confirmed before play begins.
_Avoid_: live mode, destructive mode, real mode

**Field guide**:
The in-game reference, opened from mode select, that explains each archetype, tiers, Bolts and
Symptoms. It reads no cluster data, so its examples are illustrative.
_Avoid_: info screen, help, manual, codex

**Chaos playground**:
A disposable local Kubernetes environment containing safe Pods for trying chaos mode without
targeting production infrastructure.
_Avoid_: dummy cluster, test cluster, fake cluster
