package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/davidharvith/argos/alerter"
	"github.com/davidharvith/argos/analyzer"
	"github.com/davidharvith/argos/ingestor"
	"github.com/davidharvith/argos/parser"
)

const (
	// Channel buffer sizes
	ingestBufferSize  = 1000
	parseBufferSize   = 1000
	alertBufferSize   = 100
	
	// Server ports
	httpPort = "8080"
	tcpPort  = "9090"
	
	// Worker configuration
	parserWorkers = 4
	
	// Output configuration
	alertOutputFile = "alerts.json"
)

func main() {
	log.Println("Starting Argos - Real-time Log Anomaly Detector")
	
	// Create buffered channels for data flow pipeline
	ingestChan := make(chan ingestor.LogEntry, ingestBufferSize)
	parseChan := make(chan parser.ParsedLog, parseBufferSize)
	alertChan := make(chan analyzer.Alert, alertBufferSize)
	
	// Initialize components
	ing := ingestor.NewIngestor(ingestChan, httpPort, tcpPort)
	prs := parser.NewParser(ingestChan, parseChan, parserWorkers)
	anl := analyzer.NewAnalyzer(parseChan, alertChan)
	alt := alerter.NewAlerter(alertChan, alertOutputFile)
	
	// Start all components
	if err := ing.Start(); err != nil {
		log.Fatalf("Failed to start ingestor: %v", err)
	}
	
	prs.Start()
	anl.Start()
	
	if err := alt.Start(); err != nil {
		log.Fatalf("Failed to start alerter: %v", err)
	}
	
	log.Println("Argos is running. Press Ctrl+C to stop.")
	log.Printf("HTTP endpoint: http://localhost:%s/logs", httpPort)
	log.Printf("TCP endpoint: localhost:%s", tcpPort)
	log.Printf("Alerts output: %s", alertOutputFile)
	
	// Wait for shutdown signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan
	
	log.Println("\nShutting down gracefully...")
	
	// Stop components in reverse order
	ing.Stop()
	close(ingestChan)
	
	prs.Stop()
	close(parseChan)
	
	anl.Stop()
	close(alertChan)
	
	alt.Stop()
	
	log.Println("Argos stopped successfully")
}
