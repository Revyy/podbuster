package cluster

import "context"

// EventType classifies a pod watch event.
type EventType string

const (
	EventAdded    EventType = "ADDED"
	EventDeleted  EventType = "DELETED"
	EventModified EventType = "MODIFIED"
	// EventSynced marks the end of the initial pod listing. It carries no pod.
	EventSynced EventType = "SYNCED"
)

// PodInfo carries the subset of pod data the game cares about.
type PodInfo struct {
	Name        string
	Namespace   string
	Phase       string // "Running", "Pending", "Failed", etc.
	Terminating bool   // true after deletion was requested, while the pod still exists.
	OwnerKind   string // kind of the controller owner; empty means unmanaged.
	// CPURequestMilli and MemoryRequestBytes are the effective requests summed
	// over regular and sidecar containers; zero means no request.
	CPURequestMilli    int64
	MemoryRequestBytes int64
	// Ready is true when the pod's Ready condition is True. Restarts is the
	// restart count summed over regular and sidecar containers; ordinary init
	// containers are excluded. Both are cosmetic Symptoms in the game.
	Ready    bool
	Restarts int
}

// PodEvent is emitted by a PodProvider watcher. EventModified may carry a
// Terminating pod; EventDeleted means the pod has actually been removed.
// Providers report every pod, including terminal and unmanaged ones.
type PodEvent struct {
	Type EventType
	Pod  PodInfo
}

// PodProvider abstracts cluster access. Implement this to add new data sources.
type PodProvider interface {
	// ListPods returns every pod in the given namespace.
	ListPods(namespace string) ([]PodInfo, error)
	// Watch streams pod events for the given namespace until ctx is cancelled.
	// Watch returns promptly; the initial pods arrive as EventAdded events,
	// followed by one EventSynced. A pod entering termination is reported as
	// EventModified before its final EventDeleted event.
	Watch(ctx context.Context, namespace string) (<-chan PodEvent, error)
	// DeletePod deletes the named pod. Called only in chaos mode.
	DeletePod(ctx context.Context, namespace, name string) error
}
