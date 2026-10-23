package cluster_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/revyy/podbuster/cluster"
)

func TestMockProvider_ListPods(t *testing.T) {
	p := cluster.NewMockProvider(3)
	pods, err := p.ListPods("default")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pods) != 3 {
		t.Fatalf("expected 3 pods, got %d", len(pods))
	}
}

func TestMockProvider_Watch_ReceivesInitialPods(t *testing.T) {
	p := cluster.NewMockProvider(2)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	ch, err := p.Watch(ctx, "default")
	if err != nil {
		t.Fatalf("Watch error: %v", err)
	}

	received := 0
	for received < 2 {
		select {
		case ev := <-ch:
			if ev.Type != cluster.EventAdded {
				t.Fatalf("expected EventAdded, got %s", ev.Type)
			}
			received++
		case <-ctx.Done():
			t.Fatalf("timed out waiting for events, got %d/2", received)
		}
	}
}

func TestMockProvider_Watch_ReceivesAddedPod(t *testing.T) {
	p := cluster.NewMockProvider(0)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	ch, err := p.Watch(ctx, "default")
	if err != nil {
		t.Fatalf("Watch error: %v", err)
	}

	assertEventType(t, ch, cluster.EventSynced)
	newPod := cluster.PodInfo{Name: "injected-pod", Namespace: "default", Phase: "Running"}
	p.AddPod(newPod)

	select {
	case ev := <-ch:
		if ev.Type != cluster.EventAdded || ev.Pod.Name != "injected-pod" {
			t.Fatalf("unexpected event: %+v", ev)
		}
	case <-ctx.Done():
		t.Fatal("timed out waiting for injected pod event")
	}
}

func TestMockProvider_DeletePod_EmitsTerminatingThenDeletedEvents(t *testing.T) {
	p := cluster.NewMockProvider(1)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	pods, _ := p.ListPods("default")
	podName := pods[0].Name

	ch, _ := p.Watch(ctx, "default")
	<-ch // initial ADDED
	<-ch // SYNCED

	started := time.Now()
	if err := p.DeletePod(ctx, "default", podName); err != nil {
		t.Fatalf("DeletePod error: %v", err)
	}
	if elapsed := time.Since(started); elapsed >= mockOperationReturnBudget {
		t.Fatalf("DeletePod did not return promptly: %s", elapsed)
	}

	assertPodEvent(t, ch, cluster.EventModified, podName, true)
	assertPodEvent(t, ch, cluster.EventDeleted, podName, true)
}

func TestMockProvider_TerminatePod_EmitsTerminatingThenDeletedEvents(t *testing.T) {
	p := cluster.NewMockProvider(1)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	pods, _ := p.ListPods("default")
	podName := pods[0].Name

	ch, _ := p.Watch(ctx, "default")
	<-ch // initial ADDED
	<-ch // SYNCED

	started := time.Now()
	if err := p.TerminatePod("default", podName); err != nil {
		t.Fatalf("TerminatePod error: %v", err)
	}
	if elapsed := time.Since(started); elapsed >= mockOperationReturnBudget {
		t.Fatalf("TerminatePod did not return promptly: %s", elapsed)
	}

	assertPodEvent(t, ch, cluster.EventModified, podName, true)
	assertPodEvent(t, ch, cluster.EventDeleted, podName, true)
}

func TestMockProvider_RemovePod_RemainsImmediate(t *testing.T) {
	p := cluster.NewMockProvider(1)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	pods, _ := p.ListPods("default")
	podName := pods[0].Name

	ch, _ := p.Watch(ctx, "default")
	<-ch // initial ADDED
	<-ch // SYNCED

	p.RemovePod("default", podName)
	assertPodEvent(t, ch, cluster.EventDeleted, podName, false)
}

func TestMockProvider_Watch_SyncsAfterSeedEvents(t *testing.T) {
	p := cluster.NewMockProvider(3)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	ch, err := p.Watch(ctx, "default")
	if err != nil {
		t.Fatalf("Watch error: %v", err)
	}
	for range 3 {
		assertEventType(t, ch, cluster.EventAdded)
	}
	assertEventType(t, ch, cluster.EventSynced)
}

func TestMockProvider_DefaultPodsAreOwnedByReplicaSet(t *testing.T) {
	pods, _ := cluster.NewMockProvider(2).ListPods("default")
	for _, pod := range pods {
		if pod.OwnerKind != "ReplicaSet" {
			t.Errorf("pod %s owner kind = %q, want ReplicaSet", pod.Name, pod.OwnerKind)
		}
	}
}

func assertEventType(t *testing.T, events <-chan cluster.PodEvent, eventType cluster.EventType) {
	t.Helper()
	select {
	case ev := <-events:
		if ev.Type != eventType {
			t.Fatalf("expected %s, got %+v", eventType, ev)
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("timed out waiting for %s event", eventType)
	}
}

func assertPodEvent(t *testing.T, events <-chan cluster.PodEvent, eventType cluster.EventType, podName string, terminating bool) {
	t.Helper()
	select {
	case ev := <-events:
		if ev.Type != eventType || ev.Pod.Name != podName || ev.Pod.Terminating != terminating {
			t.Fatalf("unexpected event: %+v", ev)
		}
	case <-time.After(5 * time.Second):
		t.Fatalf("timed out waiting for %s event", eventType)
	}
}

const mockOperationReturnBudget = 500 * time.Millisecond

func TestMockProvider_ReplacementAfterDeletion(t *testing.T) {
	seed := cluster.MockVarietyPods("default")
	m := cluster.NewMockProviderWithPods(seed...)
	m.EnableReplacements(50 * time.Millisecond)

	if err := m.DeletePod(context.Background(), "default", "web-c"); err != nil {
		t.Fatal(err)
	}
	if err := m.DeletePod(context.Background(), "default", "stray"); err != nil {
		t.Fatal(err)
	}

	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		pods, _ := m.ListPods("default")
		var repl *cluster.PodInfo
		names := map[string]bool{}
		for i, p := range pods {
			names[p.Name] = true
			if strings.HasPrefix(p.Name, "web-c-") {
				repl = &pods[i]
			}
		}
		if repl != nil {
			if names["web-c"] || names["stray"] || repl.OwnerKind != "ReplicaSet" ||
				repl.CPURequestMilli != 500 || repl.MemoryRequestBytes != 512<<20 {
				t.Fatalf("bad replacement state: %+v names=%v", repl, names)
			}
			time.Sleep(200 * time.Millisecond)
			pods, _ = m.ListPods("default")
			for _, p := range pods {
				if strings.HasPrefix(p.Name, "stray") {
					t.Fatalf("stray was recreated: %+v", p)
				}
			}
			return
		}
		time.Sleep(25 * time.Millisecond)
	}
	t.Fatal("no replacement appeared")
}

func TestMockProvider_NoReplacementByDefault(t *testing.T) {
	m := cluster.NewMockProvider(1)
	_ = m.TerminatePod("default", "mock-pod-0")
	time.Sleep(2*time.Second + 300*time.Millisecond)
	if pods, _ := m.ListPods("default"); len(pods) != 0 {
		t.Fatalf("unexpected pods: %+v", pods)
	}
}

func TestMockVarietyPods_SeedSymptoms(t *testing.T) {
	want := map[string]struct {
		ready    bool
		restarts int
	}{"web-a": {true, 0}, "web-c": {true, 2}, "agent-y": {false, 0}, "batch-2": {true, 6}}
	for _, p := range cluster.MockVarietyPods("default") {
		if w, ok := want[p.Name]; ok && (p.Ready != w.ready || p.Restarts != w.restarts) {
			t.Errorf("%s: got ready=%v restarts=%d", p.Name, p.Ready, p.Restarts)
		}
	}
}

func TestMockProvider_SetSymptomsEmitsModified(t *testing.T) {
	m := cluster.NewMockProviderWithPods(cluster.MockVarietyPods("default")...)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ch, _ := m.Watch(ctx, "default")
	for ev := range ch {
		if ev.Type == cluster.EventSynced {
			break
		}
	}
	if !m.SetSymptoms("default", "web-a", false, 3) {
		t.Fatal("SetSymptoms reported a missing pod")
	}
	select {
	case ev := <-ch:
		if ev.Type != cluster.EventModified || ev.Pod.Name != "web-a" || ev.Pod.Ready || ev.Pod.Restarts != 3 {
			t.Fatalf("unexpected event: %+v", ev)
		}
	case <-time.After(time.Second):
		t.Fatal("no Modified event")
	}
	if m.SetSymptoms("default", "nope", true, 0) {
		t.Fatal("expected false for a missing pod")
	}
}

func TestMockProvider_ReplacementBootsNotReady(t *testing.T) {
	m := cluster.NewMockProviderWithPods(cluster.MockVarietyPods("default")...)
	m.EnableReplacements(20 * time.Millisecond)
	m.SetBootDelay(150 * time.Millisecond)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ch, _ := m.Watch(ctx, "default")
	for ev := range ch {
		if ev.Type == cluster.EventSynced {
			break
		}
	}
	_ = m.TerminatePod("default", "web-c")
	var name string
	timeout := time.After(6 * time.Second)
	for {
		select {
		case ev := <-ch:
			switch {
			case ev.Type == cluster.EventAdded && strings.HasPrefix(ev.Pod.Name, "web-c-"):
				if ev.Pod.Ready || ev.Pod.Restarts != 0 {
					t.Fatalf("replacement must start not ready with 0 restarts: %+v", ev.Pod)
				}
				name = ev.Pod.Name
			case ev.Type == cluster.EventModified && ev.Pod.Name == name && name != "":
				if !ev.Pod.Ready {
					t.Fatalf("expected the replacement to become ready: %+v", ev.Pod)
				}
				return
			}
		case <-timeout:
			t.Fatal("replacement never became ready")
		}
	}
}

func TestMockProvider_SimulateActivity(t *testing.T) {
	m := cluster.NewMockProviderWithPods(cluster.MockVarietyPods("default")...)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go m.SimulateActivity(ctx, "default", "agent-y", 30*time.Millisecond, "web-c", 30*time.Millisecond)
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		pods, _ := m.ListPods("default")
		for _, p := range pods {
			if p.Name == "web-c" && p.Restarts > 2 {
				return
			}
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("web-c never gained a restart")
}

func TestMockVarietyPods_UsesRequestedNamespace(t *testing.T) {
	m := cluster.NewMockProviderWithPods(cluster.MockVarietyPods("prod")...)
	if pods, _ := m.ListPods("default"); len(pods) != 0 {
		t.Fatalf("expected no pods in default, got %d", len(pods))
	}
	if pods, _ := m.ListPods("prod"); len(pods) != len(cluster.MockVarietyPods("prod")) {
		t.Fatalf("expected the seed in prod, got %d", len(pods))
	}
}
