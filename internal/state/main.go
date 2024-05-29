package state

import (
	"errors"
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

type Provider interface {
	Snapshot() ([]byte, error)
}

type Persistor interface {
	Save(stateProvider Provider) error
	Load() ([]byte, error)
}

type AutosavingPersistor interface {
	Persistor
	Start(stateProvider Provider) error
	Stop() error
}

type persistor struct {
	filePath string

	autosaveInterval time.Duration
	stopChan         chan struct{}

	autosaveStarted bool
	wg              *sync.WaitGroup
}

// NewPersistor creates a state persistor that saves to a file.
func NewPersistor(filePath string) Persistor {
	return persistor{
		filePath: filePath,
	}
}

// NewPersistorWithAutosave creates a state persistor that autosaves the state to a file.
func NewAutosavingPersistor(filePath string, autosaveInterval time.Duration) AutosavingPersistor {
	stopChan := make(chan struct{})
	autosavingStarted := false
	p := persistor{
		filePath,
		autosaveInterval,
		stopChan,
		autosavingStarted,
		&sync.WaitGroup{},
	}
	return &p
}

func (p persistor) Save(stateProvider Provider) error {
	state, err := stateProvider.Snapshot()
	if err != nil {
		return fmt.Errorf("unable to get the state snapshot: %w", err)
	}

	f, err := os.Create(p.filePath)
	if err != nil {
		return fmt.Errorf("unable to create or open the state file for writing: %w", err)
	}

	_, err = f.Write(state)
	if err != nil {
		return fmt.Errorf("unable to write to the state file: %w", err)
	}

	return nil
}

func (p persistor) Load() ([]byte, error) {
	state, err := os.ReadFile(p.filePath)
	if err != nil {
		return nil, fmt.Errorf("unable to read the state file: %w", err)
	}

	return state, nil
}

func (p *persistor) Start(stateProvider Provider) error {
	if p.autosaveStarted {
		return errors.New("autosaving has already started")
	}
	p.wg.Add(1)
	go p.autoSave(stateProvider, p.autosaveInterval)
	p.autosaveStarted = true
	return nil
}

func (p persistor) Stop() error {
	if !p.autosaveStarted {
		return errors.New("autosaving is not started")
	}
	p.stopChan <- struct{}{}

	close(p.stopChan)
	p.wg.Wait()
	log.Println("autosaving stopped")
	return nil
}

func (p persistor) autoSave(stateProvider Provider, interval time.Duration) {
	ticker := time.NewTicker(interval)
mainLoop:
	for {
		select {
		case _, ok := <-ticker.C:
			if !ok {
				continue
			}
			if err := p.Save(stateProvider); err != nil {
				log.Printf("%s", fmt.Errorf("unable to autosave to file: %w", err))
				continue
			}
			log.Printf("autosaved to file")
		case _, ok := <-p.stopChan:
			if ok {
				if err := p.Save(stateProvider); err != nil {
					log.Printf("%s", fmt.Errorf("unable to save to file: %w", err))
				} else {
					log.Printf("saved to file")
				}
				break mainLoop
			}
		}
	}

	ticker.Stop()
	p.wg.Done()
}
