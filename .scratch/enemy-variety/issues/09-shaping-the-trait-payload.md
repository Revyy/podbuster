# Shaping the trait payload

Type: grilling
Status: resolved
Blocked by: 08

## Question

How do Pod traits cross from `cluster/` through `bridge/` to `game/` without `game/` importing either? Decide:
- what spawn events carry: the raw traits (owner kind, requests) with Archetype and Tier derived in `game/`, or Archetype and Tier derived before they cross. Banding is namespace-relative, so something has to see all known Pods.
- where the namespace median is kept, and how it is recomputed as Pods come and go
- how live overlay signals (if "Deciding live cosmetic overlays" keeps any) reach existing enemies. Today the bridge ignores Pod modifications other than terminating.
- where the single eligibility rule excluding unmanaged Pods lives so both the real and mock providers apply it, and so a future config option can include unmanaged Pods
- how `PodInfo` grows

## Answer

- **Layering**:
  - Providers report **raw traits** only.
  - The **bridge** derives Archetype (from owner kind, by the mapping in [Choosing the archetype roster](03-choosing-the-archetype-roster.md)) and Tier (from requests, by [Defining tier banding](04-defining-tier-banding.md)).
  - `game/` receives only game-world values. The Archetype and Tier enums live in `game/events.go`, and Kubernetes words never cross into `game/`.
- **`PodInfo`**: the raw fields `cluster/` adds:
  - `OwnerKind string`: the kind of the controller owner; empty means unmanaged.
  - `CPURequestMilli int64` and `MemoryRequestBytes int64`: effective requests summed as in tier banding; zero means no request.
  - `Ready bool` and `Restarts int`: as defined in [Deciding live cosmetic overlays](08-deciding-live-cosmetic-overlays.md).

  Parsing Kubernetes quantities stays in `toPodInfo`. The mock can set these fields per Pod; the exact API belongs to [Choosing mock and playground variety](10-choosing-mock-and-playground-variety.md).
- **Eligibility**:
  - One rule, `eligible(PodInfo)` in `bridge/`: managed and not in a terminal phase. It is the single home for a future "include unmanaged Pods" option.
  - Providers report all Pods. The terminal-phase filter and the terminal-to-Deleted mapping move out of `kube.go`.
  - A Modified event for an unknown Pod that is now eligible spawns it. A known Pod that becomes ineligible despawns through the existing deletion path.
- **Initial sync**:
  - `PodEvent` gains `EventSynced`. `KubeClient.Watch` returns its channel immediately and sends `EventSynced` from a goroutine once the handler registration's `HasSynced` is true. The mock sends it right after its seed events.
  - The bridge holds initial Pods until sync, bands them against the full set, then spawns them. Later Pods are banded on arrival.
  - This also removes a likely existing deadlock: today `Watch` waits for sync before anyone reads its 100-event buffer, so a namespace with more than 100 Pods would block.
- **Median**:
  - The bridge keeps a per-session map of known eligible Pods and their requests. It is updated on add, delete, ineligibility and in-place request changes, and cleared at session end.
  - The median is computed on demand for each spawn.
  - A request change never re-tiers an existing enemy.
- **`GameEvent`**:
  - Gains `Archetype`, `Tier` and `Symptoms{Ready bool; Restarts int}`.
  - `EventSpawnEnemy` fills all three.
  - The new `EventSymptomsChanged` carries Symptoms only. It is emitted on a Modified event only when Ready or Restarts changed.
  - The game bands restarts into patches and detects increases for the reboot blink.
- **Practice respawn**: the game's pending respawn record keeps Archetype and Tier. `EventSymptomsChanged` updates the record's Symptoms while the enemy is away. The bridge never re-sends spawns.
- **Docs to update during implementation**: the `AGENTS.md` sections on `cluster/` and `bridge/`, which describe filtering in `kube.go` and ignoring other modifications.
