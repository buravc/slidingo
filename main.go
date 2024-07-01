package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	filePath := flag.String("save", "/tmp/requestcounter.json", "file path the state will be saved to")
	addr := flag.String("addr", ":3000", "address the server listens")
	autosaveDurationStr := flag.String("autosave", "30s", "autosave interval")
	counterWindowStr := flag.String("window", "60s", "window of the request counter")

	maxConcurReq := flag.Int("maxCon", 5, "maximum number of concurrent requests")
	timeoutStr := flag.String("timeout", "300ms", "timeout window")
	flag.Parse()

	autosaveDuration, err := time.ParseDuration(*autosaveDurationStr)
	if err != nil {
		panic(fmt.Errorf("invalid duration is passed as an argument: %w", err))
	}

	counterWindow, err := time.ParseDuration(*counterWindowStr)
	if err != nil {
		panic(fmt.Errorf("invalid window is passed as an argument: %w", err))
	}

	timeout, err := time.ParseDuration(*timeoutStr)
	if err != nil {
		panic(fmt.Errorf("invalid timeout is passed as an argument: %w", err))
	}

	app := NewApp(*filePath, *addr, *maxConcurReq, timeout, autosaveDuration, counterWindow)

	go ListenGracefulShutdown(app)
	if err := app.Start(); err != nil {
		log.Printf("app shutting down: %v", err)
	}
}

func ListenGracefulShutdown(app *App) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	if err := app.Stop(); err != nil {
		panic(fmt.Errorf("unable to shutdown server: %w", err))
	}
}
