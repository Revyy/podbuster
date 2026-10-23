package main

import (
	"context"
	"flag"
	"log"
	"time"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/revyy/podbuster/bridge"
	"github.com/revyy/podbuster/cluster"
	"github.com/revyy/podbuster/game"
)

func main() {
	// There is deliberately no --chaos flag: chaos mode is chosen and
	// confirmed in-game. See docs/adr/0001-chaos-mode-is-an-in-game-confirmed-choice.md
	mock := flag.Bool("mock", false, "Use mock provider (no cluster required)")
	namespace := flag.String("namespace", "default", "Kubernetes namespace to watch")
	kubeconfig := flag.String("kubeconfig", "", "Path to kubeconfig file (defaults to $KUBECONFIG, ~/.kube/config, or in-cluster config)")
	kubeContext := flag.String("context", "", "Kubernetes context to use")
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var provider cluster.PodProvider
	if *mock {
		mockProvider := cluster.NewMockProviderWithPods(cluster.MockVarietyPods(*namespace)...)
		mockProvider.EnableReplacements(cluster.MockReplacementDelay)
		provider = mockProvider
		go mockProvider.SimulateActivity(ctx, *namespace, "agent-y", 10*time.Second, "web-c", 20*time.Second)
		log.Println("podbuster: using mock provider with the variety seed")
	} else {
		realClient, err := cluster.NewKubeClient(*kubeconfig, *kubeContext)
		if err != nil {
			log.Fatalf("podbuster: failed to initialize kubernetes client: %v", err)
		}
		provider = realClient
		log.Printf("podbuster: connected to real cluster watching namespace '%s'", *namespace)
	}

	b := bridge.NewBridge(provider)

	go func() {
		if err := b.Run(ctx); err != nil && err != context.Canceled {
			log.Printf("bridge error: %v", err)
		}
	}()

	g := game.NewGame(
		game.Config{Namespace: *namespace},
		b.GameEvents(),
		b.KillEvents(),
		b.SessionEvents(),
	)

	ebiten.SetWindowSize(800, 600)
	ebiten.SetWindowTitle("Podbuster 🚀")

	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
