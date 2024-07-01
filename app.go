package main

import (
	"fmt"
	"log"
	"slidingo/internal/request"
	"slidingo/internal/server"
	"slidingo/internal/state"
	"time"
)

type App struct {
	persistor state.AutosavingPersistor
	server    *server.HTTPServer
	counter   request.Counter
}

func NewApp(filePath, addr string, maxConcurrentRequestCount int, timeout, autosaveDuration, counterWindow time.Duration) *App {
	persistor := state.NewAutosavingPersistor(filePath, autosaveDuration)
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

	httpServer := server.New(addr, maxConcurrentRequestCount, timeout, counter)

	return &App{
		counter:   counter,
		persistor: persistor,
		server:    httpServer,
	}
}

func NewAppWithZeroState(filePath, addr string, maxConcurrentRequestCount int, timeout, autosaveDuration, counterWindow time.Duration) *App {
	persistor := state.NewAutosavingPersistor(filePath, autosaveDuration)
	counter := request.NewCounter(counterWindow)

	httpServer := server.New(addr, maxConcurrentRequestCount, timeout, counter)

	return &App{
		counter:   counter,
		persistor: persistor,
		server:    httpServer,
	}
}

func (a *App) Start() error {
	a.persistor.Start(a.counter)
	errorChan := make(chan error)
	go func() {
		if err := a.server.Start(); err != nil {
			log.Printf("application is shutting down: %v", err)
			errorChan <- err
			return
		}
		errorChan <- nil
	}()

	err := <-errorChan
	close(errorChan)
	return err
}

func (a *App) Stop() error {
	if err := a.server.Close(); err != nil {
		return err
	}
	if err := a.persistor.Stop(); err != nil {
		return err
	}

	return nil
}

func (s *App) SetTimeout(timeout time.Duration) {
	s.server.SetTimeout(timeout)
}

func (s *App) Clear() {
	s.counter.Clear()
}
