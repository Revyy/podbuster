package bridge_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/revyy/podbuster/bridge"
	"github.com/revyy/podbuster/cluster"
	"github.com/revyy/podbuster/game"
)

// testSessionID is the session the startSession helper opens.
const testSessionID = 1

// startSession boots a bridge and opens a session, returning the bridge.
func startSession(t *testing.T, ctx context.Context, provider cluster.PodProvider, chaosMode bool) *bridge.Bridge {
	t.Helper()
	b := bridge.NewBridge(provider)
	go b.Run(ctx) //nolint:errcheck
	b.SessionEvents() <- game.SessionEvent{
		Type:      game.EventSessionStart,
		ID:        testSessionID,
		Namespace: "default",
		ChaosMode: chaosMode,
	}
	return b
}

func expectEvent(t *testing.T, ctx context.Context, b *bridge.Bridge) game.GameEvent {
	t.Helper()
	select {
	case ev := <-b.GameEvents():
		return ev
	case <-ctx.Done():
		t.Fatal("timed out waiting for game event")
		return game.GameEvent{}
	}
}

func TestBridge_SpawnEnemyOnPodAdded(t *testing.T) {
	provider := cluster.NewMockProvider(0)
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()

	b := startSession(t, ctx, provider, false)

	// Give the watch a moment to register before adding a pod.
	time.Sleep(50 * time.Millisecond)
	provider.AddPod(cluster.PodInfo{Name: "test-pod", Namespace: "default", Phase: "Running", OwnerKind: "ReplicaSet"})

	ev := expectEvent(t, ctx, b)
	if ev.Type != game.EventSpawnEnemy || ev.PodName != "test-pod" {
		t.Fatalf("unexpected event: %+v", ev)
	}
}

func TestBridge_DespawnEnemyOnPodDeleted(t *testing.T) {
	provider := cluster.NewMockProvider(1)
	pods, _ := provider.ListPods("default")
	podName := pods[0].Name

	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()

	b := startSession(t, ctx, provider, false)

	if ev := expectEvent(t, ctx, b); ev.Type != game.EventSpawnEnemy {
		t.Fatalf("expected initial spawn, got %+v", ev)
	}

	provider.DeletePod(ctx, "default", podName) //nolint:errcheck

	ev := expectEvent(t, ctx, b)
	if ev.Type != game.EventPodTerminating || ev.PodName != podName {
		t.Fatalf("unexpected terminating event: %+v", ev)
	}
	ev = expectEvent(t, ctx, b)
	if ev.Type != game.EventDespawnEnemy || ev.PodName != podName {
		t.Fatalf("unexpected event: %+v", ev)
	}
}

func TestBridge_TerminatingPodEmitsTerminatingEvent(t *testing.T) {
	provider := cluster.NewMockProvider(1)
	pods, _ := provider.ListPods("default")
	podName := pods[0].Name
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	b := startSession(t, ctx, provider, false)
	expectEvent(t, ctx, b) // initial spawn

	if err := provider.TerminatePod("default", podName); err != nil {
		t.Fatalf("TerminatePod error: %v", err)
	}

	ev := expectEvent(t, ctx, b)
	if ev.Type != game.EventPodTerminating || ev.PodName != podName || ev.Namespace != "default" {
		t.Fatalf("unexpected event: %+v", ev)
	}
}

func TestBridge_DeletionFailureEmitsFailureEvent(t *testing.T) {
	provider := &failingDeleteProvider{PodProvider: cluster.NewMockProvider(1)}
	pods, _ := provider.ListPods("default")
	podName := pods[0].Name
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	b := startSession(t, ctx, provider, true)
	expectEvent(t, ctx, b) // initial spawn

	b.KillEvents() <- game.KillEvent{PodName: podName, Namespace: "default", SessionID: testSessionID}
	ev := expectEvent(t, ctx, b)
	if ev.Type != game.EventDeletionFailed || ev.PodName != podName || ev.Namespace != "default" {
		t.Fatalf("unexpected event: %+v", ev)
	}
	select {
	case ev := <-b.GameEvents():
		if ev.Type == game.EventDeletionFailed {
			t.Fatalf("expected exactly one deletion failure event, got %+v", ev)
		}
	case <-time.After(100 * time.Millisecond):
	}
}

func TestBridge_NonTerminatingModificationIsIgnored(t *testing.T) {
	provider := &modifiedProvider{pod: cluster.PodInfo{Name: "modified-pod", Namespace: "default", Phase: "Running", OwnerKind: "ReplicaSet"}}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	b := startSession(t, ctx, provider, false)
	if ev := expectEvent(t, ctx, b); ev.Type != game.EventSpawnEnemy {
		t.Fatalf("expected initial spawn, got %+v", ev)
	}
	select {
	case ev := <-b.GameEvents():
		t.Fatalf("expected non-terminating modification to be ignored, got %+v", ev)
	case <-time.After(100 * time.Millisecond):
	}
}

func TestBridge_ChaosMode_DeletesPodOnKill(t *testing.T) {
	provider := cluster.NewMockProvider(1)
	pods, _ := provider.ListPods("default")
	podName := pods[0].Name

	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()

	b := startSession(t, ctx, provider, true)
	expectEvent(t, ctx, b) // initial spawn

	b.KillEvents() <- game.KillEvent{PodName: podName, Namespace: "default", SessionID: testSessionID}
	if ev := expectEvent(t, ctx, b); ev.Type != game.EventPodTerminating {
		t.Fatalf("expected terminating event after successful deletion request, got %+v", ev)
	}

	time.Sleep(100 * time.Millisecond)
	remaining, _ := provider.ListPods("default")
	if len(remaining) != 1 || !remaining[0].Terminating {
		t.Fatalf("expected one terminating pod after kill in chaos mode, got %+v", remaining)
	}

	select {
	case <-time.After(2200 * time.Millisecond):
	case <-ctx.Done():
		t.Fatal("timed out waiting for mock deletion")
	}
	remaining, _ = provider.ListPods("default")
	if len(remaining) != 0 {
		t.Fatalf("expected 0 pods after termination delay, got %d", len(remaining))
	}
}

func TestBridge_PracticeMode_LeavesPodAlone(t *testing.T) {
	provider := cluster.NewMockProvider(1)
	pods, _ := provider.ListPods("default")
	podName := pods[0].Name

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	b := startSession(t, ctx, provider, false)
	expectEvent(t, ctx, b) // initial spawn

	b.KillEvents() <- game.KillEvent{PodName: podName, Namespace: "default", SessionID: testSessionID}

	time.Sleep(100 * time.Millisecond)
	remaining, _ := provider.ListPods("default")
	if len(remaining) != 1 {
		t.Fatalf("expected pod to survive a practice-mode kill, got %d pods", len(remaining))
	}
}

// A kill arriving with no session running must never reach the cluster.
func TestBridge_KillOutsideSessionDeletesNothing(t *testing.T) {
	provider := cluster.NewMockProvider(1)
	pods, _ := provider.ListPods("default")
	podName := pods[0].Name

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	b := bridge.NewBridge(provider)
	go b.Run(ctx) //nolint:errcheck

	b.KillEvents() <- game.KillEvent{PodName: podName, Namespace: "default"}

	time.Sleep(100 * time.Millisecond)
	remaining, _ := provider.ListPods("default")
	if len(remaining) != 1 {
		t.Fatalf("expected pod to survive a kill with no session, got %d pods", len(remaining))
	}
}

// Ending a session must stop deletions, even in a mode that was chaos.
func TestBridge_SessionEndStopsChaosDeletions(t *testing.T) {
	provider := cluster.NewMockProvider(1)
	pods, _ := provider.ListPods("default")
	podName := pods[0].Name

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	b := startSession(t, ctx, provider, true)
	expectEvent(t, ctx, b) // initial spawn

	b.SessionEvents() <- game.SessionEvent{Type: game.EventSessionEnd}
	time.Sleep(50 * time.Millisecond)

	b.KillEvents() <- game.KillEvent{PodName: podName, Namespace: "default", SessionID: testSessionID}
	time.Sleep(100 * time.Millisecond)

	remaining, _ := provider.ListPods("default")
	if len(remaining) != 1 {
		t.Fatalf("expected pod to survive a kill after session end, got %d pods", len(remaining))
	}
}

// No cluster read may happen before a session starts.
func TestBridge_NoWatchBeforeSessionStart(t *testing.T) {
	provider := &countingProvider{PodProvider: cluster.NewMockProvider(1)}

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	b := bridge.NewBridge(provider)
	go b.Run(ctx) //nolint:errcheck

	time.Sleep(100 * time.Millisecond)
	if calls := provider.calls(); calls != 0 {
		t.Fatalf("expected no watch before session start, got %d", calls)
	}

	b.SessionEvents() <- game.SessionEvent{Type: game.EventSessionStart, Namespace: "default"}
	time.Sleep(100 * time.Millisecond)
	if calls := provider.calls(); calls != 1 {
		t.Fatalf("expected exactly 1 watch after session start, got %d", calls)
	}
}

// A failing watch must surface an error and leave the bridge usable.
func TestBridge_WatchFailureEmitsErrorAndStaysAlive(t *testing.T) {
	provider := &failingWatchProvider{PodProvider: cluster.NewMockProvider(1), fail: true}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	b := bridge.NewBridge(provider)
	go b.Run(ctx) //nolint:errcheck

	b.SessionEvents() <- game.SessionEvent{Type: game.EventSessionStart, Namespace: "default"}

	ev := expectEvent(t, ctx, b)
	if ev.Type != game.EventError {
		t.Fatalf("expected an error event, got %+v", ev)
	}
	if ev.Message == "" {
		t.Fatal("expected the error event to carry a message")
	}

	// The bridge must still serve a later session.
	provider.setFail(false)
	b.SessionEvents() <- game.SessionEvent{Type: game.EventSessionStart, Namespace: "default"}
	if ev := expectEvent(t, ctx, b); ev.Type != game.EventSpawnEnemy {
		t.Fatalf("expected the bridge to serve a later session, got %+v", ev)
	}
}

// scriptedProvider replays a fixed list of events, then idles until cancelled.
type scriptedProvider struct {
	cluster.PodProvider
	events chan cluster.PodEvent
}

func newScriptedProvider() *scriptedProvider {
	return &scriptedProvider{PodProvider: cluster.NewMockProvider(0), events: make(chan cluster.PodEvent, 16)}
}

func (s *scriptedProvider) Watch(ctx context.Context, _ string) (<-chan cluster.PodEvent, error) {
	return s.events, nil
}

func (s *scriptedProvider) send(t cluster.EventType, pod cluster.PodInfo) {
	s.events <- cluster.PodEvent{Type: t, Pod: pod}
}

func managedPod(name string) cluster.PodInfo {
	return cluster.PodInfo{Name: name, Namespace: "default", Phase: "Running", OwnerKind: "ReplicaSet"}
}

func expectNoEvent(t *testing.T, b *bridge.Bridge) {
	t.Helper()
	select {
	case ev := <-b.GameEvents():
		t.Fatalf("unexpected event: %+v", ev)
	case <-time.After(100 * time.Millisecond):
	}
}

func TestBridge_UnmanagedPodNeverSpawns(t *testing.T) {
	provider := newScriptedProvider()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	b := startSession(t, ctx, provider, false)

	unmanaged := managedPod("bare")
	unmanaged.OwnerKind = ""
	provider.send(cluster.EventAdded, unmanaged)
	provider.send(cluster.EventSynced, cluster.PodInfo{})
	provider.send(cluster.EventAdded, unmanaged)
	provider.send(cluster.EventModified, unmanaged)
	provider.send(cluster.EventDeleted, unmanaged)
	expectNoEvent(t, b)
}

func TestBridge_HoldsInitialSpawnsUntilSynced(t *testing.T) {
	provider := newScriptedProvider()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	b := startSession(t, ctx, provider, false)

	provider.send(cluster.EventAdded, managedPod("a"))
	provider.send(cluster.EventAdded, managedPod("b"))
	expectNoEvent(t, b)

	provider.send(cluster.EventSynced, cluster.PodInfo{})
	for _, name := range []string{"a", "b"} {
		if ev := expectEvent(t, ctx, b); ev.Type != game.EventSpawnEnemy || ev.PodName != name {
			t.Fatalf("expected spawn of %s, got %+v", name, ev)
		}
	}
}

func TestBridge_PodDeletedBeforeSyncNeverSpawns(t *testing.T) {
	provider := newScriptedProvider()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	b := startSession(t, ctx, provider, false)

	provider.send(cluster.EventAdded, managedPod("gone"))
	provider.send(cluster.EventDeleted, managedPod("gone"))
	provider.send(cluster.EventSynced, cluster.PodInfo{})
	expectNoEvent(t, b)
}

func TestBridge_PodReAddedBeforeSyncSpawnsOnce(t *testing.T) {
	provider := newScriptedProvider()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	b := startSession(t, ctx, provider, false)

	provider.send(cluster.EventAdded, managedPod("again"))
	provider.send(cluster.EventDeleted, managedPod("again"))
	provider.send(cluster.EventAdded, managedPod("again"))
	provider.send(cluster.EventSynced, cluster.PodInfo{})
	if ev := expectEvent(t, ctx, b); ev.Type != game.EventSpawnEnemy || ev.PodName != "again" {
		t.Fatalf("expected spawn, got %+v", ev)
	}
	expectNoEvent(t, b)
}

func TestBridge_KnownPodTurningTerminalDespawns(t *testing.T) {
	provider := newScriptedProvider()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	b := startSession(t, ctx, provider, false)

	provider.send(cluster.EventAdded, managedPod("p"))
	provider.send(cluster.EventSynced, cluster.PodInfo{})
	expectEvent(t, ctx, b) // spawn

	done := managedPod("p")
	done.Phase = "Succeeded"
	provider.send(cluster.EventModified, done)
	if ev := expectEvent(t, ctx, b); ev.Type != game.EventDespawnEnemy || ev.PodName != "p" {
		t.Fatalf("expected despawn, got %+v", ev)
	}

	// The later real deletion must not despawn a second time.
	provider.send(cluster.EventDeleted, done)
	expectNoEvent(t, b)
}

func TestBridge_UnknownPodBecomingEligibleSpawnsOnModified(t *testing.T) {
	provider := newScriptedProvider()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	b := startSession(t, ctx, provider, false)

	provider.send(cluster.EventSynced, cluster.PodInfo{})
	pending := managedPod("late")
	pending.Phase = "Failed"
	provider.send(cluster.EventAdded, pending)
	expectNoEvent(t, b)

	provider.send(cluster.EventModified, managedPod("late"))
	if ev := expectEvent(t, ctx, b); ev.Type != game.EventSpawnEnemy || ev.PodName != "late" {
		t.Fatalf("expected spawn, got %+v", ev)
	}
}

type countingProvider struct {
	cluster.PodProvider
	mu         sync.Mutex
	watchCalls int
}

func (c *countingProvider) Watch(ctx context.Context, namespace string) (<-chan cluster.PodEvent, error) {
	c.mu.Lock()
	c.watchCalls++
	c.mu.Unlock()
	return c.PodProvider.Watch(ctx, namespace)
}

func (c *countingProvider) calls() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.watchCalls
}

type failingWatchProvider struct {
	cluster.PodProvider
	mu   sync.Mutex
	fail bool
}

type failingDeleteProvider struct {
	cluster.PodProvider
}

func (f *failingDeleteProvider) DeletePod(context.Context, string, string) error {
	return errors.New("delete denied")
}

type modifiedProvider struct {
	pod cluster.PodInfo
}

func (m *modifiedProvider) ListPods(string) ([]cluster.PodInfo, error) {
	return []cluster.PodInfo{m.pod}, nil
}

func (m *modifiedProvider) Watch(ctx context.Context, _ string) (<-chan cluster.PodEvent, error) {
	events := make(chan cluster.PodEvent, 3)
	events <- cluster.PodEvent{Type: cluster.EventAdded, Pod: m.pod}
	events <- cluster.PodEvent{Type: cluster.EventSynced}
	events <- cluster.PodEvent{Type: cluster.EventModified, Pod: m.pod}
	go func() {
		<-ctx.Done()
		close(events)
	}()
	return events, nil
}

func (m *modifiedProvider) DeletePod(context.Context, string, string) error {
	return nil
}

func (f *failingWatchProvider) Watch(ctx context.Context, namespace string) (<-chan cluster.PodEvent, error) {
	f.mu.Lock()
	fail := f.fail
	f.mu.Unlock()
	if fail {
		return nil, errors.New("connection refused")
	}
	return f.PodProvider.Watch(ctx, namespace)
}

func (f *failingWatchProvider) setFail(fail bool) {
	f.mu.Lock()
	f.fail = fail
	f.mu.Unlock()
}

// A kill made in practice mode must never delete a Pod, even if it is still in
// flight when the player switches to a chaos session.
func TestBridge_StalePracticeKillDoesNotDeleteInLaterChaosSession(t *testing.T) {
	provider := cluster.NewMockProvider(1)
	pods, _ := provider.ListPods("default")
	podName := pods[0].Name

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	b := bridge.NewBridge(provider)
	go b.Run(ctx) //nolint:errcheck

	// Play a practice session, leave it, then start a chaos session. The
	// practice kill only reaches the bridge once chaos is live.
	b.SessionEvents() <- game.SessionEvent{Type: game.EventSessionStart, ID: 1, Namespace: "default", ChaosMode: false}
	time.Sleep(50 * time.Millisecond)
	b.SessionEvents() <- game.SessionEvent{Type: game.EventSessionStart, ID: 2, Namespace: "default", ChaosMode: true}
	time.Sleep(50 * time.Millisecond)

	b.KillEvents() <- game.KillEvent{PodName: podName, Namespace: "default", SessionID: 1}
	time.Sleep(150 * time.Millisecond)

	remaining, _ := provider.ListPods("default")
	if len(remaining) != 1 {
		t.Fatalf("expected a practice-mode kill never to delete a Pod, got %d pods", len(remaining))
	}
}

// A kill stamped with an ended session must be discarded, even though the
// session running when it arrives is also a chaos session.
func TestBridge_KillFromEndedSessionIsDiscarded(t *testing.T) {
	provider := cluster.NewMockProvider(1)
	pods, _ := provider.ListPods("default")
	podName := pods[0].Name

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	b := bridge.NewBridge(provider)
	go b.Run(ctx) //nolint:errcheck

	b.SessionEvents() <- game.SessionEvent{Type: game.EventSessionStart, ID: 1, Namespace: "default", ChaosMode: true}
	time.Sleep(50 * time.Millisecond)
	b.SessionEvents() <- game.SessionEvent{Type: game.EventSessionEnd}
	time.Sleep(50 * time.Millisecond)
	b.SessionEvents() <- game.SessionEvent{Type: game.EventSessionStart, ID: 2, Namespace: "default", ChaosMode: true}
	time.Sleep(50 * time.Millisecond)

	// This kill belongs to session 1, which is over.
	b.KillEvents() <- game.KillEvent{PodName: podName, Namespace: "default", SessionID: 1}
	time.Sleep(150 * time.Millisecond)

	remaining, _ := provider.ListPods("default")
	if len(remaining) != 1 {
		t.Fatalf("expected a kill from an ended session to be discarded, got %d pods", len(remaining))
	}
}

func sizedPod(name, ownerKind string, cpuMilli, memBytes int64) cluster.PodInfo {
	pod := managedPod(name)
	pod.OwnerKind = ownerKind
	pod.CPURequestMilli = cpuMilli
	pod.MemoryRequestBytes = memBytes
	return pod
}

const mi = 1024 * 1024

// spawnsByName reads n spawn events and indexes them by pod name.
func spawnsByName(t *testing.T, ctx context.Context, b *bridge.Bridge, n int) map[string]game.GameEvent {
	t.Helper()
	got := map[string]game.GameEvent{}
	for range n {
		ev := expectEvent(t, ctx, b)
		if ev.Type != game.EventSpawnEnemy {
			t.Fatalf("expected spawn, got %+v", ev)
		}
		got[ev.PodName] = ev
	}
	return got
}

func TestBridge_MapsOwnerKindToArchetype(t *testing.T) {
	provider := newScriptedProvider()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	b := startSession(t, ctx, provider, false)

	want := map[string]game.Archetype{
		"ReplicaSet":  game.ArchetypeProbe,
		"CronThing":   game.ArchetypeProbe,
		"StatefulSet": game.ArchetypeTurret,
		"DaemonSet":   game.ArchetypeSentinel,
		"Job":         game.ArchetypeDart,
	}
	for kind := range want {
		provider.send(cluster.EventAdded, sizedPod(kind, kind, 0, 0))
	}
	provider.send(cluster.EventSynced, cluster.PodInfo{})
	for name, ev := range spawnsByName(t, ctx, b, len(want)) {
		if ev.Archetype != want[name] {
			t.Errorf("%s: archetype %v, want %v", name, ev.Archetype, want[name])
		}
	}
}

func TestBridge_TierBanding(t *testing.T) {
	tests := []struct {
		name string
		pods []cluster.PodInfo
		want map[string]game.Tier
	}{
		{
			name: "uniform namespace is all standard",
			pods: []cluster.PodInfo{sizedPod("a", "ReplicaSet", 100, 128*mi), sizedPod("b", "ReplicaSet", 100, 128*mi), sizedPod("c", "ReplicaSet", 100, 128*mi)},
			want: map[string]game.Tier{"a": game.TierStandard, "b": game.TierStandard, "c": game.TierStandard},
		},
		{
			name: "no requests anywhere is all light",
			pods: []cluster.PodInfo{sizedPod("a", "ReplicaSet", 0, 0), sizedPod("b", "ReplicaSet", 0, 0)},
			want: map[string]game.Tier{"a": game.TierLight, "b": game.TierLight},
		},
		{
			name: "lone pod with requests is standard",
			pods: []cluster.PodInfo{sizedPod("a", "ReplicaSet", 500, 64*mi)},
			want: map[string]game.Tier{"a": game.TierStandard},
		},
		{
			name: "pod without requests is light and leaves the median alone",
			pods: []cluster.PodInfo{sizedPod("a", "ReplicaSet", 100, 100*mi), sizedPod("b", "ReplicaSet", 0, 0), sizedPod("c", "ReplicaSet", 100, 100*mi)},
			want: map[string]game.Tier{"a": game.TierStandard, "b": game.TierLight, "c": game.TierStandard},
		},
		{
			name: "light and heavy against the median",
			pods: []cluster.PodInfo{
				sizedPod("small", "ReplicaSet", 40, 40*mi),
				sizedPod("mid1", "ReplicaSet", 100, 100*mi),
				sizedPod("mid2", "ReplicaSet", 100, 100*mi),
				sizedPod("big", "ReplicaSet", 100, 300*mi),
				sizedPod("edge-low", "ReplicaSet", 50, 50*mi),
				sizedPod("edge-high", "ReplicaSet", 200, 200*mi),
			},
			want: map[string]game.Tier{
				"small": game.TierLight, "mid1": game.TierStandard, "mid2": game.TierStandard,
				"big": game.TierHeavy, "edge-low": game.TierStandard, "edge-high": game.TierStandard,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := newScriptedProvider()
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			b := startSession(t, ctx, provider, false)
			for _, pod := range tt.pods {
				provider.send(cluster.EventAdded, pod)
			}
			provider.send(cluster.EventSynced, cluster.PodInfo{})
			for name, ev := range spawnsByName(t, ctx, b, len(tt.pods)) {
				if ev.Tier != tt.want[name] {
					t.Errorf("%s: tier %v, want %v", name, ev.Tier, tt.want[name])
				}
			}
		})
	}
}

func TestBridge_InitialPodsBandedTogetherRegardlessOfArrivalOrder(t *testing.T) {
	provider := newScriptedProvider()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	b := startSession(t, ctx, provider, false)

	// Judged on arrival alone, "first" would be a lone, standard Pod.
	provider.send(cluster.EventAdded, sizedPod("first", "ReplicaSet", 1000, 100*mi))
	provider.send(cluster.EventAdded, sizedPod("second", "ReplicaSet", 100, 100*mi))
	provider.send(cluster.EventAdded, sizedPod("third", "ReplicaSet", 100, 100*mi))
	provider.send(cluster.EventSynced, cluster.PodInfo{})
	got := spawnsByName(t, ctx, b, 3)
	if got["first"].Tier != game.TierHeavy || got["second"].Tier != game.TierStandard {
		t.Fatalf("unexpected tiers: first=%v second=%v", got["first"].Tier, got["second"].Tier)
	}
}

func TestBridge_LaterPodBandedOnArrivalAgainstKnownPods(t *testing.T) {
	provider := newScriptedProvider()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	b := startSession(t, ctx, provider, false)

	provider.send(cluster.EventAdded, sizedPod("a", "ReplicaSet", 100, 100*mi))
	provider.send(cluster.EventAdded, sizedPod("b", "ReplicaSet", 100, 100*mi))
	provider.send(cluster.EventSynced, cluster.PodInfo{})
	spawnsByName(t, ctx, b, 2)

	provider.send(cluster.EventAdded, sizedPod("huge", "ReplicaSet", 1000, 100*mi))
	if ev := spawnsByName(t, ctx, b, 1)["huge"]; ev.Tier != game.TierHeavy {
		t.Fatalf("expected heavy, got %v", ev.Tier)
	}
}

func TestBridge_MedianFollowsDeletesAndRequestChanges(t *testing.T) {
	provider := newScriptedProvider()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	b := startSession(t, ctx, provider, false)

	small1, small2 := sizedPod("s1", "ReplicaSet", 100, 100*mi), sizedPod("s2", "ReplicaSet", 100, 100*mi)
	big1, big2 := sizedPod("b1", "ReplicaSet", 1000, 100*mi), sizedPod("b2", "ReplicaSet", 1000, 100*mi)
	for _, pod := range []cluster.PodInfo{small1, small2, big1, big2} {
		provider.send(cluster.EventAdded, pod)
	}
	provider.send(cluster.EventSynced, cluster.PodInfo{})
	spawnsByName(t, ctx, b, 4)

	// Median is 550m; 1100m is exactly 2x, so Standard.
	provider.send(cluster.EventAdded, sizedPod("probe1", "ReplicaSet", 1100, 100*mi))
	if ev := spawnsByName(t, ctx, b, 1)["probe1"]; ev.Tier != game.TierStandard {
		t.Fatalf("probe1: expected standard, got %v", ev.Tier)
	}

	// Deleting both small Pods lifts the median to 1000m; 300m is then Light.
	provider.send(cluster.EventDeleted, small1)
	provider.send(cluster.EventDeleted, small2)
	for range 2 {
		if ev := expectEvent(t, ctx, b); ev.Type != game.EventDespawnEnemy {
			t.Fatalf("expected despawn, got %+v", ev)
		}
	}
	provider.send(cluster.EventAdded, sizedPod("probe2", "ReplicaSet", 300, 30*mi))
	if ev := spawnsByName(t, ctx, b, 1)["probe2"]; ev.Tier != game.TierLight {
		t.Fatalf("probe2: expected light, got %v", ev.Tier)
	}

	// A request change moves the median but never re-tiers a live enemy.
	big1.CPURequestMilli = 10
	provider.send(cluster.EventModified, big1)
	expectNoEvent(t, b)
	provider.send(cluster.EventAdded, sizedPod("probe3", "ReplicaSet", 300, 100*mi))
	if ev := spawnsByName(t, ctx, b, 1)["probe3"]; ev.Tier != game.TierStandard {
		t.Fatalf("probe3: expected standard against the lowered median, got %v", ev.Tier)
	}
}

func TestBridge_MockVarietySeedArchetypeAndTierMix(t *testing.T) {
	provider := cluster.NewMockProviderWithPods(cluster.MockVarietyPods("default")...)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	b := startSession(t, ctx, provider, false)

	type trait struct {
		a game.Archetype
		t game.Tier
	}
	want := map[string]trait{
		"web-a":   {game.ArchetypeProbe, game.TierLight},
		"web-b":   {game.ArchetypeProbe, game.TierStandard},
		"web-c":   {game.ArchetypeProbe, game.TierHeavy},
		"db-0":    {game.ArchetypeTurret, game.TierStandard},
		"db-1":    {game.ArchetypeTurret, game.TierHeavy},
		"agent-x": {game.ArchetypeSentinel, game.TierLight},
		"agent-y": {game.ArchetypeSentinel, game.TierStandard},
		"batch-1": {game.ArchetypeDart, game.TierLight},
		"batch-2": {game.ArchetypeDart, game.TierStandard},
	}
	got := spawnsByName(t, ctx, b, len(want))
	for name, w := range want {
		ev, ok := got[name]
		if !ok {
			t.Errorf("%s never spawned", name)
			continue
		}
		if ev.Archetype != w.a || ev.Tier != w.t {
			t.Errorf("%s: got %v/%v, want %v/%v", name, ev.Archetype, ev.Tier, w.a, w.t)
		}
	}
	if _, ok := got["stray"]; ok {
		t.Error("unmanaged stray spawned")
	}
	select {
	case ev := <-b.GameEvents():
		if ev.Type == game.EventSpawnEnemy {
			t.Fatalf("unexpected extra spawn: %+v", ev)
		}
	case <-time.After(100 * time.Millisecond):
	}
}

func symptomPod(name string, ready bool, restarts int) cluster.PodInfo {
	p := managedPod(name)
	p.Ready, p.Restarts = ready, restarts
	return p
}

func TestBridge_SpawnCarriesSymptoms(t *testing.T) {
	provider := newScriptedProvider()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	b := startSession(t, ctx, provider, false)

	provider.send(cluster.EventAdded, symptomPod("early", false, 6))
	provider.send(cluster.EventSynced, cluster.PodInfo{})
	provider.send(cluster.EventAdded, symptomPod("late", true, 1))

	want := map[string]game.Symptoms{"early": {Ready: false, Restarts: 6}, "late": {Ready: true, Restarts: 1}}
	for range 2 {
		ev := expectEvent(t, ctx, b)
		if ev.Type != game.EventSpawnEnemy || ev.Symptoms != want[ev.PodName] {
			t.Fatalf("unexpected spawn %+v", ev)
		}
	}
}

func TestBridge_SymptomsChangedOnlyWhenReadyOrRestartsChange(t *testing.T) {
	provider := newScriptedProvider()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	b := startSession(t, ctx, provider, false)

	provider.send(cluster.EventAdded, symptomPod("a", true, 0))
	provider.send(cluster.EventSynced, cluster.PodInfo{})
	expectEvent(t, ctx, b)

	unrelated := symptomPod("a", true, 0)
	unrelated.CPURequestMilli = 500
	provider.send(cluster.EventModified, unrelated)
	provider.send(cluster.EventModified, unrelated)
	expectNoEvent(t, b)

	provider.send(cluster.EventModified, symptomPod("a", false, 0))
	ev := expectEvent(t, ctx, b)
	if ev.Type != game.EventSymptomsChanged || ev.PodName != "a" || ev.Symptoms != (game.Symptoms{Ready: false, Restarts: 0}) {
		t.Fatalf("unexpected event %+v", ev)
	}
	provider.send(cluster.EventModified, symptomPod("a", false, 3))
	ev = expectEvent(t, ctx, b)
	if ev.Type != game.EventSymptomsChanged || ev.Symptoms != (game.Symptoms{Ready: false, Restarts: 3}) {
		t.Fatalf("unexpected event %+v", ev)
	}
	provider.send(cluster.EventModified, symptomPod("a", false, 3))
	expectNoEvent(t, b)
}

func TestBridge_SymptomChangeBeforeSyncFoldsIntoSpawn(t *testing.T) {
	provider := newScriptedProvider()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	b := startSession(t, ctx, provider, false)

	provider.send(cluster.EventAdded, symptomPod("a", false, 0))
	provider.send(cluster.EventModified, symptomPod("a", true, 1))
	provider.send(cluster.EventSynced, cluster.PodInfo{})
	ev := expectEvent(t, ctx, b)
	if ev.Type != game.EventSpawnEnemy || ev.Symptoms != (game.Symptoms{Ready: true, Restarts: 1}) {
		t.Fatalf("unexpected event %+v", ev)
	}
	expectNoEvent(t, b)
}

func TestBridge_MockSetSymptomsReachesGame(t *testing.T) {
	mock := cluster.NewMockProviderWithPods(symptomPod("a", true, 0))
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	b := startSession(t, ctx, mock, false)
	expectEvent(t, ctx, b)

	mock.SetSymptoms("default", "a", true, 1)
	ev := expectEvent(t, ctx, b)
	if ev.Type != game.EventSymptomsChanged || ev.Symptoms.Restarts != 1 {
		t.Fatalf("unexpected event %+v", ev)
	}
}
