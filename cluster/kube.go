package cluster

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

// KubeClient implements PodProvider using client-go.
type KubeClient struct {
	clientset kubernetes.Interface
}

// NewKubeClient creates a new KubeClient given kubeconfig and context options.
// If kubeconfig or kubeContext are empty, standard clientcmd default rules apply.
func NewKubeClient(kubeconfig, kubeContext string) (*KubeClient, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	if kubeconfig != "" {
		loadingRules.ExplicitPath = kubeconfig
	}
	configOverrides := &clientcmd.ConfigOverrides{}
	if kubeContext != "" {
		configOverrides.CurrentContext = kubeContext
	}

	clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
	restConfig, err := clientConfig.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load kubernetes config: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes clientset: %w", err)
	}

	return &KubeClient{clientset: clientset}, nil
}

// NewKubeClientWithInterface creates a KubeClient with a given kubernetes.Interface.
func NewKubeClientWithInterface(clientset kubernetes.Interface) *KubeClient {
	return &KubeClient{clientset: clientset}
}

// ListPods returns every pod in the given namespace, whatever its phase or owner.
func (k *KubeClient) ListPods(namespace string) ([]PodInfo, error) {
	pods, err := k.clientset.CoreV1().Pods(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list pods in namespace %s: %w", namespace, err)
	}

	var result []PodInfo
	for _, pod := range pods.Items {
		result = append(result, toPodInfo(&pod))
	}
	return result, nil
}

// Watch streams pod events for the given namespace until ctx is cancelled.
func (k *KubeClient) Watch(ctx context.Context, namespace string) (<-chan PodEvent, error) {
	factory := informers.NewSharedInformerFactoryWithOptions(
		k.clientset,
		0,
		informers.WithNamespace(namespace),
	)

	podInformer := factory.Core().V1().Pods().Informer()
	ch := make(chan PodEvent, 100)

	registration, err := podInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			pod, ok := obj.(*corev1.Pod)
			if !ok {
				return
			}
			select {
			case ch <- PodEvent{Type: EventAdded, Pod: toPodInfo(pod)}:
			case <-ctx.Done():
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			pod, ok := newObj.(*corev1.Pod)
			if !ok {
				return
			}
			select {
			case ch <- PodEvent{Type: EventModified, Pod: toPodInfo(pod)}:
			case <-ctx.Done():
			}
		},
		DeleteFunc: func(obj interface{}) {
			pod, ok := obj.(*corev1.Pod)
			if !ok {
				tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
				if !ok {
					return
				}
				pod, ok = tombstone.Obj.(*corev1.Pod)
				if !ok {
					return
				}
			}
			select {
			case ch <- PodEvent{Type: EventDeleted, Pod: toPodInfo(pod)}:
			case <-ctx.Done():
			}
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to add event handler: %w", err)
	}

	factory.Start(ctx.Done())

	go func() {
		if cache.WaitForCacheSync(ctx.Done(), registration.HasSynced) {
			select {
			case ch <- PodEvent{Type: EventSynced}:
			case <-ctx.Done():
			}
		}
		<-ctx.Done()
		close(ch)
	}()

	return ch, nil
}

// DeletePod deletes the named pod in the specified namespace.
func (k *KubeClient) DeletePod(ctx context.Context, namespace, name string) error {
	err := k.clientset.CoreV1().Pods(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete pod %s in namespace %s: %w", name, namespace, err)
	}
	return nil
}

func toPodInfo(pod *corev1.Pod) PodInfo {
	ownerKind := ""
	if owner := metav1.GetControllerOf(pod); owner != nil {
		ownerKind = owner.Kind
	}
	cpu, mem := effectiveRequests(pod)
	return PodInfo{
		CPURequestMilli:    cpu,
		MemoryRequestBytes: mem,
		OwnerKind:          ownerKind,
		Name:               pod.Name,
		Namespace:          pod.Namespace,
		Phase:              string(pod.Status.Phase),
		Terminating:        pod.DeletionTimestamp != nil,
		Ready:              podReady(pod),
		Restarts:           restarts(pod),
	}
}

func podReady(pod *corev1.Pod) bool {
	for _, c := range pod.Status.Conditions {
		if c.Type == corev1.PodReady {
			return c.Status == corev1.ConditionTrue
		}
	}
	return false
}

// restarts sums restartCount over regular containers and sidecars (restartable
// init containers). Ordinary init containers are ignored, like in
// effectiveRequests.
func restarts(pod *corev1.Pod) int {
	sidecars := map[string]bool{}
	for _, c := range pod.Spec.InitContainers {
		if c.RestartPolicy != nil && *c.RestartPolicy == corev1.ContainerRestartPolicyAlways {
			sidecars[c.Name] = true
		}
	}
	total := 0
	for _, st := range pod.Status.ContainerStatuses {
		total += int(st.RestartCount)
	}
	for _, st := range pod.Status.InitContainerStatuses {
		if sidecars[st.Name] {
			total += int(st.RestartCount)
		}
	}
	return total
}

// effectiveRequests sums CPU (millicores) and memory (bytes) requests over
// regular containers and sidecars (restartable init containers). Ordinary init
// containers run before the rest and are ignored. Limits are ignored too: the
// API server already copies a lone limit into the request.
func effectiveRequests(pod *corev1.Pod) (cpuMilli, memBytes int64) {
	add := func(c *corev1.Container) {
		cpuMilli += c.Resources.Requests.Cpu().MilliValue()
		memBytes += c.Resources.Requests.Memory().Value()
	}
	for i := range pod.Spec.Containers {
		add(&pod.Spec.Containers[i])
	}
	for i := range pod.Spec.InitContainers {
		c := &pod.Spec.InitContainers[i]
		if c.RestartPolicy != nil && *c.RestartPolicy == corev1.ContainerRestartPolicyAlways {
			add(c)
		}
	}
	return cpuMilli, memBytes
}
