package cluster

import (
	"context"
	"fmt"
	"sync"
	"time"
)

const mockTerminationDelay = 2 * time.Second

type mockSubscriber struct {
	namespace string
	ch        chan PodEvent
}

// MockProvider is an in-memory PodProvider for development and testing.
type MockProvider struct {
	mu          sync.Mutex
	pods        map[string]PodInfo
	subscribers []mockSubscriber

	// replaceDelay is zero unless replacements are enabled.
	replaceDelay time.Duration
	replaceSeq   int
	// replacementBase maps a replacement's name to the name it descends from.
	replacementBase map[string]string
	// bootDelay is how long a replacement stays not ready.
	bootDelay time.Duration
}

// MockBootDelay is how long a simulated replacement Pod stays not ready.
const MockBootDelay = 3 * time.Second

// MockReplacementDelay is how long a simulated controller takes to recreate a
// deleted Pod.
const MockReplacementDelay = 2 * time.Second

const (
	mockMi = 1 << 20
)

// MockVarietyPods returns the fixed seed used by --mock, placed in namespace:
// nine eligible Pods covering every Archetype and Tier, plus the unmanaged
// "stray" Pod, which the bridge never turns into an enemy.
func MockVarietyPods(namespace string) []PodInfo {
	pod := func(name, owner string, cpuMilli, memMi int64) PodInfo {
		return PodInfo{
			Name: name, Namespace: namespace, Phase: "Running", OwnerKind: owner,
			CPURequestMilli: cpuMilli, MemoryRequestBytes: memMi * mockMi,
			Ready: true,
		}
	}
	notReady := func(p PodInfo) PodInfo { p.Ready = false; return p }
	restarted := func(p PodInfo, n int) PodInfo { p.Restarts = n; return p }
	return []PodInfo{
		pod("web-a", "ReplicaSet", 0, 0),
		pod("web-b", "ReplicaSet", 100, 128),
		restarted(pod("web-c", "ReplicaSet", 500, 512), 2),
		pod("db-0", "StatefulSet", 100, 128),
		pod("db-1", "StatefulSet", 500, 512),
		pod("agent-x", "DaemonSet", 25, 32),
		notReady(pod("agent-y", "DaemonSet", 100, 128)),
		pod("batch-1", "Job", 25, 32),
		restarted(pod("batch-2", "Job", 100, 128), 6),
		pod("stray", "", 0, 0),
	}
}

// NewMockProviderWithPods creates a MockProvider seeded with exactly the
// given pods.
func NewMockProviderWithPods(pods ...PodInfo) *MockProvider {
	m := &MockProvider{
		pods:            make(map[string]PodInfo),
		replacementBase: make(map[string]string),
	}
	for _, pod := range pods {
		m.pods[podKey(pod.Namespace, pod.Name)] = pod
	}
	return m
}

// EnableReplacements makes the mock simulate controllers: delay after a
// managed pod's deletion completes, a replacement with the same owner kind
// and requests but a new name suffix is added. Unmanaged pods are never
// replaced.
func (m *MockProvider) EnableReplacements(delay time.Duration) {
	m.mu.Lock()
	m.replaceDelay = delay
	m.bootDelay = MockBootDelay
	m.mu.Unlock()
}

// SetBootDelay changes how long replacements stay not ready.
func (m *MockProvider) SetBootDelay(d time.Duration) {
	m.mu.Lock()
	m.bootDelay = d
	m.mu.Unlock()
}

// SetSymptoms sets a pod's readiness and restart count and broadcasts an
// EventModified. It reports whether the pod exists.
func (m *MockProvider) SetSymptoms(namespace, name string, ready bool, restarts int) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.setSymptomsLocked(podKey(namespace, name), ready, restarts)
}

func (m *MockProvider) setSymptomsLocked(key string, ready bool, restarts int) bool {
	pod, ok := m.pods[key]
	if !ok {
		return false
	}
	pod.Ready, pod.Restarts = ready, restarts
	m.pods[key] = pod
	m.broadcast(PodEvent{Type: EventModified, Pod: pod})
	return true
}

// SimulateActivity runs the --mock ticker until ctx is cancelled: every
// readyEvery it flips the readiness of namespace/readyPod, and every
// restartEvery it adds a restart to namespace/restartPod.
func (m *MockProvider) SimulateActivity(ctx context.Context, namespace, readyPod string, readyEvery time.Duration, restartPod string, restartEvery time.Duration) {
	flip := time.NewTicker(readyEvery)
	restart := time.NewTicker(restartEvery)
	defer flip.Stop()
	defer restart.Stop()
	for {
		select {
		case <-flip.C:
			m.mu.Lock()
			key := podKey(namespace, readyPod)
			if pod, ok := m.pods[key]; ok {
				m.setSymptomsLocked(key, !pod.Ready, pod.Restarts)
			}
			m.mu.Unlock()
		case <-restart.C:
			m.mu.Lock()
			key := podKey(namespace, restartPod)
			if pod, ok := m.pods[key]; ok {
				m.setSymptomsLocked(key, pod.Ready, pod.Restarts+1)
			}
			m.mu.Unlock()
		case <-ctx.Done():
			return
		}
	}
}

// scheduleReplacement must be called with m.mu held.
func (m *MockProvider) scheduleReplacement(deleted PodInfo) {
	if m.replaceDelay <= 0 || deleted.OwnerKind == "" {
		return
	}
	base, ok := m.replacementBase[deleted.Name]
	if !ok {
		base = deleted.Name
	}
	time.AfterFunc(m.replaceDelay, func() {
		m.mu.Lock()
		m.replaceSeq++
		repl := deleted
		repl.Name = fmt.Sprintf("%s-r%d", base, m.replaceSeq)
		repl.Terminating = false
		repl.Phase = "Running"
		repl.Restarts = 0
		repl.Ready = m.bootDelay <= 0
		m.replacementBase[repl.Name] = base
		m.pods[podKey(repl.Namespace, repl.Name)] = repl
		m.broadcast(PodEvent{Type: EventAdded, Pod: repl})
		replKey := podKey(repl.Namespace, repl.Name)
		if boot := m.bootDelay; boot > 0 {
			time.AfterFunc(boot, func() {
				m.mu.Lock()
				defer m.mu.Unlock()
				if cur, ok := m.pods[replKey]; ok && !cur.Terminating {
					m.setSymptomsLocked(replKey, true, cur.Restarts)
				}
			})
		}
		m.mu.Unlock()
	})
}

// NewMockProvider creates a MockProvider pre-populated with count fake pods
// all in the "default" namespace.
func NewMockProvider(count int) *MockProvider {
	pods := make([]PodInfo, 0, count)
	for i := range count {
		pods = append(pods, PodInfo{
			Name:      fmt.Sprintf("mock-pod-%d", i),
			Namespace: "default",
			Phase:     "Running",
			OwnerKind: "ReplicaSet",
			Ready:     true,
		})
	}
	return NewMockProviderWithPods(pods...)
}

func podKey(namespace, name string) string {
	return namespace + "/" + name
}

func (m *MockProvider) addSubscriber(namespace string, ch chan PodEvent) {
	m.subscribers = append(m.subscribers, mockSubscriber{namespace: namespace, ch: ch})
}

func (m *MockProvider) removeSubscriber(ch chan PodEvent) {
	for i, sub := range m.subscribers {
		if sub.ch == ch {
			m.subscribers = append(m.subscribers[:i], m.subscribers[i+1:]...)
			return
		}
	}
}

func (m *MockProvider) broadcast(event PodEvent) {
	for _, sub := range m.subscribers {
		if sub.namespace != event.Pod.Namespace {
			continue
		}
		select {
		case sub.ch <- event:
		default:
		}
	}
}

// ListPods returns all pods in the given namespace.
func (m *MockProvider) ListPods(namespace string) ([]PodInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var result []PodInfo
	for _, p := range m.pods {
		if p.Namespace == namespace {
			result = append(result, p)
		}
	}
	return result, nil
}

// Watch returns a channel that receives PodEvents. It first emits EventAdded
// for all existing pods, then EventSynced, then streams live changes until ctx is cancelled.
func (m *MockProvider) Watch(ctx context.Context, namespace string) (<-chan PodEvent, error) {
	ch := make(chan PodEvent, 64)

	m.mu.Lock()
	m.addSubscriber(namespace, ch)
	existing := make([]PodInfo, 0)
	for _, p := range m.pods {
		if p.Namespace == namespace {
			existing = append(existing, p)
		}
	}
	m.mu.Unlock()

	go func() {
		defer func() {
			m.mu.Lock()
			m.removeSubscriber(ch)
			m.mu.Unlock()
			close(ch)
		}()

		for _, p := range existing {
			select {
			case ch <- PodEvent{Type: EventAdded, Pod: p}:
			case <-ctx.Done():
				return
			}
		}

		select {
		case ch <- PodEvent{Type: EventSynced}:
		case <-ctx.Done():
			return
		}

		<-ctx.Done()
	}()

	return ch, nil
}

// AddPod adds a pod and broadcasts an EventAdded to matching watchers.
func (m *MockProvider) AddPod(pod PodInfo) {
	m.mu.Lock()
	m.pods[podKey(pod.Namespace, pod.Name)] = pod
	m.broadcast(PodEvent{Type: EventAdded, Pod: pod})
	m.mu.Unlock()
}

// RemovePod removes a pod and broadcasts EventDeleted to matching watchers.
func (m *MockProvider) RemovePod(namespace, name string) {
	m.mu.Lock()
	key := podKey(namespace, name)
	pod, ok := m.pods[key]
	if ok {
		delete(m.pods, key)
		m.broadcast(PodEvent{Type: EventDeleted, Pod: pod})
	}
	m.mu.Unlock()
}

// DeletePod implements PodProvider. It marks the pod terminating immediately
// and removes it after the mock termination delay.
func (m *MockProvider) DeletePod(_ context.Context, namespace, name string) error {
	return m.terminatePod(namespace, name)
}

// TerminatePod simulates an externally-caused pod termination.
func (m *MockProvider) TerminatePod(namespace, name string) error {
	return m.terminatePod(namespace, name)
}

func (m *MockProvider) terminatePod(namespace, name string) error {
	m.mu.Lock()
	key := podKey(namespace, name)
	pod, ok := m.pods[key]
	if !ok {
		m.mu.Unlock()
		return fmt.Errorf("pod %s/%s not found", namespace, name)
	}

	if pod.Terminating {
		m.mu.Unlock()
		return nil
	}

	pod.Terminating = true
	m.pods[key] = pod
	m.broadcast(PodEvent{Type: EventModified, Pod: pod})
	m.mu.Unlock()

	go func() {
		time.Sleep(mockTerminationDelay)
		m.mu.Lock()
		defer m.mu.Unlock()
		current, ok := m.pods[key]
		if !ok || !current.Terminating {
			return
		}
		delete(m.pods, key)
		m.broadcast(PodEvent{Type: EventDeleted, Pod: current})
		m.scheduleReplacement(current)
	}()

	return nil
}
