package game

// GameEventType classifies events sent from the bridge into the game.
type GameEventType string

const (
	// EventSpawnEnemy instructs the scene to create a new enemy for a pod.
	EventSpawnEnemy GameEventType = "SPAWN"
	// EventDespawnEnemy instructs the scene to remove the enemy for a pod.
	EventDespawnEnemy GameEventType = "DESPAWN"
	// EventPodTerminating tells the scene that a pod is entering termination.
	EventPodTerminating GameEventType = "POD_TERMINATING"
	// EventDeletionFailed tells the scene that a requested pod deletion failed.
	EventDeletionFailed GameEventType = "DELETION_FAILED"
	// EventSymptomsChanged tells the scene that a pod's Symptoms changed.
	EventSymptomsChanged GameEventType = "SYMPTOMS_CHANGED"
	// EventError reports that the bridge could not serve the session.
	// The game surfaces Message to the player and refuses to start play.
	EventError GameEventType = "ERROR"
)

// Archetype is the kind of droid an enemy is, fixed at spawn.
type Archetype int

const (
	ArchetypeProbe Archetype = iota
	ArchetypeTurret
	ArchetypeSentinel
	ArchetypeDart
)

// Tier is how heavy an enemy is relative to its neighbours, fixed at spawn.
type Tier int

const (
	TierLight Tier = iota
	TierStandard
	TierHeavy
)

// Symptoms are a pod's live, purely cosmetic signals. They never affect
// behaviour or stats.
type Symptoms struct {
	Ready    bool
	Restarts int
}

// GameEvent is sent by the bridge to the game loop over a channel.
// The game package defines this type; the bridge package imports it.
type GameEvent struct {
	Type      GameEventType
	PodName   string
	Namespace string
	Message   string
	// Archetype and Tier are set on EventSpawnEnemy.
	Archetype Archetype
	Tier      Tier
	// Symptoms is set on EventSpawnEnemy and EventSymptomsChanged.
	Symptoms Symptoms
}

// KillEvent is sent by the game to the bridge when a player destroys an enemy.
// The bridge will optionally call DeletePod in chaos mode.
//
// SessionID identifies the session the kill happened in. The bridge discards
// kills from any other session, so a kill made in practice mode can never be
// carried over into a later chaos session and delete a real Pod.
type KillEvent struct {
	PodName   string
	Namespace string
	SessionID uint64
}

// SessionEventType classifies session lifecycle events sent from the game to
// the bridge.
type SessionEventType string

const (
	// EventSessionStart asks the bridge to open a watch for a namespace.
	EventSessionStart SessionEventType = "SESSION_START"
	// EventSessionEnd asks the bridge to close the current watch and idle.
	EventSessionEnd SessionEventType = "SESSION_END"
)

// SessionEvent is sent by the game to the bridge to start or end a session.
// A session owns exactly one namespace watch; the bridge reads no cluster data
// outside one. ChaosMode travels with the start event so that the mode and the
// watch can never disagree.
//
// ID identifies the session. The game stamps every KillEvent with it so that a
// kill can only ever take effect in the session that produced it.
type SessionEvent struct {
	Type      SessionEventType
	ID        uint64
	Namespace string
	ChaosMode bool
}
