package cluster

import (
	"context"
	"fmt"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
)

func TestKubeClientListPods(t *testing.T) {
	pod1 := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "pod-1", Namespace: "default"},
		Status:     corev1.PodStatus{Phase: corev1.PodRunning},
	}
	pod2 := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "pod-2", Namespace: "default"},
		Status:     corev1.PodStatus{Phase: corev1.PodSucceeded},
	}
	pod3 := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "pod-3", Namespace: "other"},
		Status:     corev1.PodStatus{Phase: corev1.PodPending},
	}

	clientset := fake.NewSimpleClientset(pod1, pod2, pod3)
	client := NewKubeClientWithInterface(clientset)

	pods, err := client.ListPods("default")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(pods) != 2 {
		t.Fatalf("expected both default-namespace pods, terminal included, got %d", len(pods))
	}
}

func TestKubeClientDeletePod(t *testing.T) {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "pod-to-delete", Namespace: "default"},
		Status:     corev1.PodStatus{Phase: corev1.PodRunning},
	}

	clientset := fake.NewSimpleClientset(pod)
	client := NewKubeClientWithInterface(clientset)

	err := client.DeletePod(context.Background(), "default", "pod-to-delete")
	if err != nil {
		t.Fatalf("unexpected error on DeletePod: %v", err)
	}

	pods, err := client.ListPods("default")
	if err != nil {
		t.Fatalf("unexpected error on ListPods: %v", err)
	}
	if len(pods) != 0 {
		t.Errorf("expected 0 pods after delete, got %d", len(pods))
	}
}

func TestKubeClientWatch(t *testing.T) {
	pod1 := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "pod-initial", Namespace: "test-ns"},
		Status:     corev1.PodStatus{Phase: corev1.PodRunning},
	}

	clientset := fake.NewSimpleClientset(pod1)
	client := NewKubeClientWithInterface(clientset)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	eventCh, err := client.Watch(ctx, "test-ns")
	if err != nil {
		t.Fatalf("failed to start Watch: %v", err)
	}

	select {
	case evt := <-eventCh:
		if evt.Type != EventAdded {
			t.Errorf("expected EventAdded, got %s", evt.Type)
		}
		if evt.Pod.Name != "pod-initial" {
			t.Errorf("expected pod-initial, got %s", evt.Pod.Name)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("timed out waiting for initial watch event")
	}

	waitForEventType(t, eventCh, EventSynced)

	// Add a new pod dynamically
	pod2 := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "pod-dynamic", Namespace: "test-ns"},
		Status:     corev1.PodStatus{Phase: corev1.PodRunning},
	}
	_, err = clientset.CoreV1().Pods("test-ns").Create(ctx, pod2, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("failed to create dynamic pod: %v", err)
	}

	select {
	case evt := <-eventCh:
		if evt.Type != EventAdded {
			t.Errorf("expected EventAdded, got %s", evt.Type)
		}
		if evt.Pod.Name != "pod-dynamic" {
			t.Errorf("expected pod-dynamic, got %s", evt.Pod.Name)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("timed out waiting for dynamic pod event")
	}
}

func TestKubeClientWatchTerminatingThenDeleted(t *testing.T) {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "pod-terminating", Namespace: "test-ns"},
		Status:     corev1.PodStatus{Phase: corev1.PodRunning},
	}

	clientset := fake.NewSimpleClientset(pod)
	client := NewKubeClientWithInterface(clientset)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	eventCh, err := client.Watch(ctx, "test-ns")
	if err != nil {
		t.Fatalf("failed to start Watch: %v", err)
	}
	waitForPodEvent(t, eventCh, EventAdded, false)

	terminating := pod.DeepCopy()
	now := metav1.Now()
	terminating.DeletionTimestamp = &now
	if _, err := clientset.CoreV1().Pods("test-ns").Update(ctx, terminating, metav1.UpdateOptions{}); err != nil {
		t.Fatalf("failed to mark pod terminating: %v", err)
	}
	waitForPodEvent(t, eventCh, EventModified, true)

	if err := clientset.CoreV1().Pods("test-ns").Delete(ctx, pod.Name, metav1.DeleteOptions{}); err != nil {
		t.Fatalf("failed to delete pod: %v", err)
	}
	waitForPodEvent(t, eventCh, EventDeleted, true)
}

func waitForPodEvent(t *testing.T, events <-chan PodEvent, eventType EventType, terminating bool) {
	t.Helper()
	select {
	case event := <-events:
		if event.Type != eventType {
			t.Fatalf("expected %s, got %s", eventType, event.Type)
		}
		if event.Pod.Terminating != terminating {
			t.Fatalf("expected Terminating=%t, got %t", terminating, event.Pod.Terminating)
		}
	case <-time.After(1 * time.Second):
		t.Fatalf("timed out waiting for %s event", eventType)
	}
}

func waitForEventType(t *testing.T, events <-chan PodEvent, eventType EventType) PodEvent {
	t.Helper()
	select {
	case event := <-events:
		if event.Type != eventType {
			t.Fatalf("expected %s, got %s", eventType, event.Type)
		}
		return event
	case <-time.After(2 * time.Second):
		t.Fatalf("timed out waiting for %s event", eventType)
		return PodEvent{}
	}
}

func TestKubeClientWatchReportsOwnerKind(t *testing.T) {
	controller := true
	managed := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: "managed", Namespace: "ns",
			OwnerReferences: []metav1.OwnerReference{
				{Kind: "ConfigMap", Name: "x"},
				{Kind: "StatefulSet", Name: "db", Controller: &controller},
			},
		},
		Status: corev1.PodStatus{Phase: corev1.PodRunning},
	}
	bare := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "bare", Namespace: "ns"},
		Status:     corev1.PodStatus{Phase: corev1.PodRunning},
	}
	client := NewKubeClientWithInterface(fake.NewSimpleClientset(managed, bare))
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	ch, err := client.Watch(ctx, "ns")
	if err != nil {
		t.Fatalf("failed to start Watch: %v", err)
	}
	kinds := map[string]string{}
	for range 2 {
		ev := waitForEventType(t, ch, EventAdded)
		kinds[ev.Pod.Name] = ev.Pod.OwnerKind
	}
	if kinds["managed"] != "StatefulSet" || kinds["bare"] != "" {
		t.Fatalf("unexpected owner kinds: %v", kinds)
	}
}

func TestKubeClientWatchReportsTerminalPodsRaw(t *testing.T) {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "p", Namespace: "ns"},
		Status:     corev1.PodStatus{Phase: corev1.PodRunning},
	}
	clientset := fake.NewSimpleClientset(pod)
	client := NewKubeClientWithInterface(clientset)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	ch, err := client.Watch(ctx, "ns")
	if err != nil {
		t.Fatalf("failed to start Watch: %v", err)
	}
	waitForEventType(t, ch, EventAdded)
	waitForEventType(t, ch, EventSynced)

	done := pod.DeepCopy()
	done.Status.Phase = corev1.PodSucceeded
	if _, err := clientset.CoreV1().Pods("ns").Update(ctx, done, metav1.UpdateOptions{}); err != nil {
		t.Fatalf("update failed: %v", err)
	}
	ev := waitForEventType(t, ch, EventModified)
	if ev.Pod.Phase != "Succeeded" {
		t.Fatalf("expected raw Succeeded phase, got %q", ev.Pod.Phase)
	}
}

func TestKubeClientWatchManyPodsDoesNotBlock(t *testing.T) {
	const count = 250
	var objs []runtime.Object
	for i := range count {
		objs = append(objs, &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf("pod-%d", i), Namespace: "ns"},
			Status:     corev1.PodStatus{Phase: corev1.PodRunning},
		})
	}
	client := NewKubeClientWithInterface(fake.NewSimpleClientset(objs...))
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	returned := make(chan (<-chan PodEvent), 1)
	go func() {
		ch, err := client.Watch(ctx, "ns")
		if err != nil {
			t.Errorf("failed to start Watch: %v", err)
		}
		returned <- ch
	}()
	var ch <-chan PodEvent
	select {
	case ch = <-returned:
	case <-time.After(time.Second):
		t.Fatal("Watch did not return promptly")
	}

	for range count {
		waitForEventType(t, ch, EventAdded)
	}
	waitForEventType(t, ch, EventSynced)
}

func requests(cpu, mem string) corev1.ResourceRequirements {
	list := corev1.ResourceList{}
	if cpu != "" {
		list[corev1.ResourceCPU] = resource.MustParse(cpu)
	}
	if mem != "" {
		list[corev1.ResourceMemory] = resource.MustParse(mem)
	}
	return corev1.ResourceRequirements{Requests: list}
}

func TestKubeClientReportsEffectiveRequests(t *testing.T) {
	always := corev1.ContainerRestartPolicyAlways
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "p", Namespace: "ns"},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{Name: "app", Resources: requests("100m", "128Mi")},
				{Name: "web", Resources: requests("250m", "")},
			},
			InitContainers: []corev1.Container{
				{Name: "sidecar", RestartPolicy: &always, Resources: requests("50m", "64Mi")},
				{Name: "setup", Resources: requests("2", "4Gi")},
			},
		},
		Status: corev1.PodStatus{Phase: corev1.PodRunning},
	}
	bare := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "bare", Namespace: "ns"},
		Spec:       corev1.PodSpec{Containers: []corev1.Container{{Name: "c"}}},
	}
	pods, err := NewKubeClientWithInterface(fake.NewSimpleClientset(pod, bare)).ListPods("ns")
	if err != nil {
		t.Fatalf("ListPods: %v", err)
	}
	got := map[string]PodInfo{}
	for _, p := range pods {
		got[p.Name] = p
	}
	if p := got["p"]; p.CPURequestMilli != 400 || p.MemoryRequestBytes != 192*1024*1024 {
		t.Fatalf("unexpected requests: cpu=%d mem=%d", p.CPURequestMilli, p.MemoryRequestBytes)
	}
	if p := got["bare"]; p.CPURequestMilli != 0 || p.MemoryRequestBytes != 0 {
		t.Fatalf("expected no requests, got %+v", p)
	}
}

func TestKubeClientReportsSymptoms(t *testing.T) {
	always := corev1.ContainerRestartPolicyAlways
	cond := func(s corev1.ConditionStatus) []corev1.PodCondition {
		return []corev1.PodCondition{{Type: corev1.PodReady, Status: s}}
	}
	busy := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "busy", Namespace: "ns"},
		Spec: corev1.PodSpec{InitContainers: []corev1.Container{
			{Name: "sidecar", RestartPolicy: &always}, {Name: "setup"},
		}},
		Status: corev1.PodStatus{
			Conditions:        cond(corev1.ConditionTrue),
			ContainerStatuses: []corev1.ContainerStatus{{Name: "a", RestartCount: 2}, {Name: "b", RestartCount: 1}},
			InitContainerStatuses: []corev1.ContainerStatus{
				{Name: "sidecar", RestartCount: 4}, {Name: "setup", RestartCount: 9},
			},
		},
	}
	booting := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "booting", Namespace: "ns"},
		Status:     corev1.PodStatus{Conditions: cond(corev1.ConditionFalse)},
	}
	bare := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "bare", Namespace: "ns"}}
	pods, err := NewKubeClientWithInterface(fake.NewSimpleClientset(busy, booting, bare)).ListPods("ns")
	if err != nil {
		t.Fatalf("ListPods: %v", err)
	}
	got := map[string]PodInfo{}
	for _, p := range pods {
		got[p.Name] = p
	}
	if p := got["busy"]; !p.Ready || p.Restarts != 7 {
		t.Fatalf("busy: want ready, 7 restarts, got %+v", p)
	}
	if p := got["booting"]; p.Ready || p.Restarts != 0 {
		t.Fatalf("booting: want not ready, got %+v", p)
	}
	if p := got["bare"]; p.Ready {
		t.Fatalf("pod without a Ready condition must be not ready, got %+v", p)
	}
}
