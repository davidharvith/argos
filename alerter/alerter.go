package alerter

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/davidharvith/argos/analyzer"
)

// Alerter handles alert output and notification
type Alerter struct {
	alertChan <-chan analyzer.Alert
	outputFile string
	file      *os.File
	mu        sync.Mutex
	shutdown  chan struct{}
	wg        sync.WaitGroup
}

// NewAlerter creates a new Alerter instance
func NewAlerter(alertChan <-chan analyzer.Alert, outputFile string) *Alerter {
	return &Alerter{
		alertChan:  alertChan,
		outputFile: outputFile,
		shutdown:   make(chan struct{}),
	}
}

// Start begins the alerter
func (a *Alerter) Start() error {
	// Open output file
	var err error
	if a.outputFile != "" {
		a.file, err = os.OpenFile(a.outputFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return fmt.Errorf("failed to open output file: %w", err)
		}
	}
	
	a.wg.Add(1)
	go a.processAlerts()
	log.Println("Alerter started")
	return nil
}

// processAlerts reads alerts and outputs them
func (a *Alerter) processAlerts() {
	defer a.wg.Done()
	
	for {
		select {
		case alert, ok := <-a.alertChan:
			if !ok {
				return
			}
			a.outputAlert(alert)
		case <-a.shutdown:
			return
		}
	}
}

// outputAlert formats and outputs an alert
func (a *Alerter) outputAlert(alert analyzer.Alert) {
	a.mu.Lock()
	defer a.mu.Unlock()
	
	alertJSON, err := json.MarshalIndent(alert, "", "  ")
	if err != nil {
		log.Printf("Failed to marshal alert: %v", err)
		return
	}
	
	// Print to console
	fmt.Printf("\n🚨 ALERT: %s (Severity: %s)\n", alert.Reason, alert.Severity)
	fmt.Println(string(alertJSON))
	fmt.Println(strings.Repeat("-", 80))
	
	// Write to file if configured
	if a.file != nil {
		a.file.Write(alertJSON)
		a.file.Write([]byte("\n"))
	}
}

// Stop gracefully shuts down the alerter
func (a *Alerter) Stop() {
	close(a.shutdown)
	a.wg.Wait()
	
	if a.file != nil {
		a.file.Close()
	}
	
	log.Println("Alerter stopped")
}
