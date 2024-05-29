package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"slidingo/internal/request"
	"slidingo/internal/server"
	"slidingo/internal/state"
	"syscall"
	"time"
)

func main() {
	filePath := flag.String("save", "/tmp/requestcounter.json", "file path the state will be saved to")
	addr := flag.String("addr", ":3000", "address the server listens")
	autosaveDurationStr := flag.String("autosave", "30s", "autosave interval")
	counterWindowStr := flag.String("window", "60s", "window of the request counter")
	flag.Parse()

	autosaveDuration, err := time.ParseDuration(*autosaveDurationStr)
	if err != nil {
		panic(fmt.Errorf("invalid duration is passed as an argument: %w", err))
	}

	counterWindow, err := time.ParseDuration(*counterWindowStr)
	if err != nil {
		panic(fmt.Errorf("invalid window is passed as an argument: %w", err))
	}

	persistor := state.NewAutosavingPersistor(*filePath, autosaveDuration)
	state, err := persistor.Load()
	var counter request.Counter
	if err != nil {
		log.Println("no previous state is found")
		counter = request.NewCounter(counterWindow)
	} else {
		log.Println("loading from previous state")
		counter, err = request.NewCounterFromSnapshot(state, counterWindow)
		if err != nil {
			panic(fmt.Errorf("unable to load from previous state: %w", err))
		}
	}

	handler := server.NewHandler(counter)

	httpServer := server.New(*addr)
	httpServer.SetHandler(handler)

	persistor.Start(counter)
	defer persistor.Stop()

	go ListenGracefulShutdown(httpServer)
	if err = httpServer.Start(); err != nil {
		log.Printf("application is shutting down: %v", err)
	}
}

func ListenGracefulShutdown(server *server.HTTPServer) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	if err := server.Close(); err != nil {
		panic(fmt.Errorf("unable to shutdown server: %w", err))
	}
}
