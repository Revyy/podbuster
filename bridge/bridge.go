package bridge

import (
	"context"
	"log"
	"slices"
	"sort"

	"github.com/revyy/podbuster/cluster"
	"github.com/revyy/podbuster/game"
)

// Bridge translates cluster PodEvents into game GameEvents and routes
// KillEvents from the game back to the cluster provider (in chaos mode).
//
// The bridge is idle until the game starts a session. A session owns exactly
// one namespace watch, so no cluster data is read while the player sits on the
// title screen.
type Bridge struct {
	provider      cluster.PodProvider
	gameEvents    chan game.GameEvent
	killEvents    chan game.KillEvent
	sessionEvents chan game.SessionEvent
}

// session holds the state of one active watch.
type session struct {
	id        uint64
	namespace string
	chaosMode bool
	podEvents <-chan cluster.PodEvent
	cancel    context.CancelFunc

	// known holds every eligible pod that has been or will be spawned.
	known  map[string]cluster.PodInfo
	synced bool
	// held lists the pods seen before the initial sync, in arrival order.
	held []string
}

// eligible is the single rule deciding whether a pod is an enemy: it must be
// managed (have a controller owner) and not be in a terminal phase.
func eligible(pod cluster.PodInfo) bool {
	if pod.OwnerKind == "" {
		return false
	}
	return pod.Phase != "Succeeded" && pod.Phase != "Failed"
}

// archetypeFor maps a controller owner kind to an Archetype. Unrecognised
// kinds are Probes, so no managed Pod goes missing.
func archetypeFor(ownerKind string) game.Archetype {
	switch ownerKind {
	case "StatefulSet":
		return game.ArchetypeTurret
	case "DaemonSet":
		return game.ArchetypeSentinel
	case "Job":
		return game.ArchetypeDart
	default:
		return game.ArchetypeProbe
	}
}

// Tier band edges, as ratios to the namespace median request.
const (
	lightBelow = 0.5
	heavyAbove = 2.0
)

// tierFor bands a pod against the median of the session's known eligible pods
// that have requests. The score is the larger of the CPU and memory ratios. A
// pod with no requests is Light and never feeds the median.
func (s *session) tierFor(pod cluster.PodInfo) game.Tier {
	if pod.CPURequestMilli <= 0 && pod.MemoryRequestBytes <= 0 {
		return game.TierLight
	}
	score := 0.0
	if cpu := median(s.known, func(p cluster.PodInfo) int64 { return p.CPURequestMilli }); cpu > 0 && pod.CPURequestMilli > 0 {
		score = max(score, float64(pod.CPURequestMilli)/cpu)
	}
	if mem := median(s.known, func(p cluster.PodInfo) int64 { return p.MemoryRequestBytes }); mem > 0 && pod.MemoryRequestBytes > 0 {
		score = max(score, float64(pod.MemoryRequestBytes)/mem)
	}
	switch {
	case score < lightBelow:
		return game.TierLight
	case score > heavyAbove:
		return game.TierHeavy
	default:
		return game.TierStandard
	}
}

// median returns the median of the positive values of one resource across the
// known pods, or 0 when none has a request.
func median(known map[string]cluster.PodInfo, value func(cluster.PodInfo) int64) float64 {
	var values []int64
	for _, pod := range known {
		if v := value(pod); v > 0 {
			values = append(values, v)
		}
	}
	if len(values) == 0 {
		return 0
	}
	sort.Slice(values, func(i, j int) bool { return values[i] < values[j] })
	mid := len(values) / 2
	if len(values)%2 == 1 {
		return float64(values[mid])
	}
	return float64(values[mid-1]+values[mid]) / 2
}

func podKey(pod cluster.PodInfo) string {
	return pod.Namespace + "/" + pod.Name
}

// NewBridge creates a Bridge. Call Run(ctx) as a goroutine to start it.
func NewBridge(provider cluster.PodProvider) *Bridge {
	return &Bridge{
		provider:      provider,
		gameEvents:    make(chan game.GameEvent, 64),
		killEvents:    make(chan game.KillEvent, 64),
		sessionEvents: make(chan game.SessionEvent, 8),
	}
}

// GameEvents returns the read-only channel of events destined for the game loop.
func (b *Bridge) GameEvents() <-chan game.GameEvent {
	return b.gameEvents
}

// KillEvents returns the write-only channel the game sends kill events to.
func (b *Bridge) KillEvents() chan<- game.KillEvent {
	return b.killEvents
}

// SessionEvents returns the write-only channel the game uses to start and end
// sessions.
func (b *Bridge) SessionEvents() chan<- game.SessionEvent {
	return b.sessionEvents
}

// Run starts the bridge. It blocks until ctx is cancelled. A failed watch ends
// the session but never the bridge: Run returns to idle so the player can try
// again from the menu.
func (b *Bridge) Run(ctx context.Context) error {
	var current *session
	defer func() {
		if current != nil {
			current.cancel()
		}
	}()

	for {
		// A nil channel blocks forever in select, which is exactly the
		// behaviour we want while no session is running.
		var podEvents <-chan cluster.PodEvent
		if current != nil {
			podEvents = current.podEvents
		}

		select {
		case ev, ok := <-podEvents:
			if !ok {
				// The watch ended on its own; drop back to idle.
				current.cancel()
				current = nil
				continue
			}
			b.handlePodEvent(current, ev)
		case kill := <-b.killEvents:
			b.handleKill(ctx, kill, current)
		case sessionEvent := <-b.sessionEvents:
			current = b.handleSessionEvent(ctx, sessionEvent, current)
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (b *Bridge) handleSessionEvent(ctx context.Context, ev game.SessionEvent, current *session) *session {
	if current != nil {
		current.cancel()
		current = nil
	}

	if ev.Type != game.EventSessionStart {
		return nil
	}

	sessionCtx, cancel := context.WithCancel(ctx)
	podEvents, err := b.provider.Watch(sessionCtx, ev.Namespace)
	if err != nil {
		cancel()
		log.Printf("bridge: failed to watch namespace %s: %v", ev.Namespace, err)
		b.emit(game.GameEvent{
			Type:      game.EventError,
			Namespace: ev.Namespace,
			Message:   err.Error(),
		})
		return nil
	}

	return &session{
		id:        ev.ID,
		namespace: ev.Namespace,
		chaosMode: ev.ChaosMode,
		podEvents: podEvents,
		cancel:    cancel,
		known:     make(map[string]cluster.PodInfo),
	}
}

func (b *Bridge) handlePodEvent(s *session, ev cluster.PodEvent) {
	if ev.Type == cluster.EventSynced {
		s.synced = true
		for _, key := range s.held {
			if pod, ok := s.known[key]; ok {
				b.emitSpawn(s, pod)
				if pod.Terminating {
					b.emitPod(game.EventPodTerminating, pod)
				}
			}
		}
		s.held = nil
		return
	}

	key := podKey(ev.Pod)
	_, isKnown := s.known[key]

	if ev.Type == cluster.EventDeleted || !eligible(ev.Pod) {
		if isKnown {
			delete(s.known, key)
			if s.synced {
				b.emitPod(game.EventDespawnEnemy, ev.Pod)
			} else {
				s.held = slices.DeleteFunc(s.held, func(k string) bool { return k == key })
			}
		}
		return
	}

	switch ev.Type {
	case cluster.EventAdded, cluster.EventModified:
	default:
		return
	}

	prev := s.known[key]
	s.known[key] = ev.Pod
	if isKnown && s.synced && ev.Type == cluster.EventModified && symptomsOf(prev) != symptomsOf(ev.Pod) {
		b.emit(game.GameEvent{
			Type:      game.EventSymptomsChanged,
			PodName:   ev.Pod.Name,
			Namespace: ev.Pod.Namespace,
			Symptoms:  symptomsOf(ev.Pod),
		})
	}
	switch {
	case !isKnown:
		if s.synced {
			b.emitSpawn(s, ev.Pod)
		} else {
			s.held = append(s.held, key)
		}
	case ev.Type == cluster.EventModified && ev.Pod.Terminating && s.synced:
		b.emitPod(game.EventPodTerminating, ev.Pod)
	}
}

func (b *Bridge) emitSpawn(s *session, pod cluster.PodInfo) {
	b.emit(game.GameEvent{
		Type:      game.EventSpawnEnemy,
		PodName:   pod.Name,
		Namespace: pod.Namespace,
		Archetype: archetypeFor(pod.OwnerKind),
		Tier:      s.tierFor(pod),
		Symptoms:  symptomsOf(pod),
	})
}

func symptomsOf(pod cluster.PodInfo) game.Symptoms {
	return game.Symptoms{Ready: pod.Ready, Restarts: pod.Restarts}
}

func (b *Bridge) emitPod(t game.GameEventType, pod cluster.PodInfo) {
	b.emit(game.GameEvent{Type: t, PodName: pod.Name, Namespace: pod.Namespace})
}

func (b *Bridge) emit(ev game.GameEvent) {
	select {
	case b.gameEvents <- ev:
	default:
		log.Printf("bridge: game event channel full, dropping %s for %s/%s", ev.Type, ev.Namespace, ev.PodName)
	}
}

// handleKill deletes the pod behind a kill, but only in a chaos-mode session.
// A kill outside a session, or in practice mode, has no cluster consequence.
//
// The kill must also belong to the session that is currently running. A kill
// can sit in the buffered channel while the player ends one session and starts
// another, and a practice-mode kill must never be re-evaluated against a later
// chaos session.
func (b *Bridge) handleKill(ctx context.Context, kill game.KillEvent, current *session) {
	if current == nil || !current.chaosMode {
		return
	}
	if kill.SessionID != current.id {
		log.Printf("bridge: discarding kill for %s/%s from stale session %d", kill.Namespace, kill.PodName, kill.SessionID)
		return
	}
	if err := b.provider.DeletePod(ctx, kill.Namespace, kill.PodName); err != nil {
		log.Printf("bridge: failed to delete pod %s/%s: %v", kill.Namespace, kill.PodName, err)
		b.emit(game.GameEvent{
			Type:      game.EventDeletionFailed,
			PodName:   kill.PodName,
			Namespace: kill.Namespace,
		})
	}
}
