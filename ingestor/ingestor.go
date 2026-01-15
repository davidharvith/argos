package ingestor

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"
)

// LogEntry represents a raw log entry received from the generator
type LogEntry struct {
	Timestamp string `json:"timestamp"`
	Level     string `json:"level"`
	Source    string `json:"source"`
	Message   string `json:"message"`
}

// Ingestor handles incoming log data via HTTP and TCP
type Ingestor struct {
	logChan    chan<- LogEntry
	httpPort   string
	tcpPort    string
	wg         sync.WaitGroup
	shutdown   chan struct{}
}

// NewIngestor creates a new Ingestor instance
func NewIngestor(logChan chan<- LogEntry, httpPort, tcpPort string) *Ingestor {
	return &Ingestor{
		logChan:  logChan,
		httpPort: httpPort,
		tcpPort:  tcpPort,
		shutdown: make(chan struct{}),
	}
}

// Start begins listening for logs on HTTP and TCP
func (i *Ingestor) Start() error {
	i.wg.Add(2)
	
	// Start HTTP server
	go i.startHTTPServer()
	
	// Start TCP server
	go i.startTCPServer()
	
	log.Println("Ingestor started on HTTP:", i.httpPort, "and TCP:", i.tcpPort)
	return nil
}

// startHTTPServer starts the HTTP log receiver
func (i *Ingestor) startHTTPServer() {
	defer i.wg.Done()
	
	mux := http.NewServeMux()
	mux.HandleFunc("/logs", i.handleHTTPLogs)
	
	server := &http.Server{
		Addr:    ":" + i.httpPort,
		Handler: mux,
	}
	
	go func() {
		<-i.shutdown
		server.Close()
	}()
	
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Printf("HTTP server error: %v", err)
	}
}

// handleHTTPLogs processes HTTP POST requests with log data
func (i *Ingestor) handleHTTPLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	var entry LogEntry
	if err := json.NewDecoder(r.Body).Decode(&entry); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	
	select {
	case i.logChan <- entry:
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Log received")
	case <-i.shutdown:
		http.Error(w, "Service shutting down", http.StatusServiceUnavailable)
	}
}

// startTCPServer starts the TCP log receiver
func (i *Ingestor) startTCPServer() {
	defer i.wg.Done()
	
	listener, err := net.Listen("tcp", ":"+i.tcpPort)
	if err != nil {
		log.Printf("TCP server error: %v", err)
		return
	}
	defer listener.Close()
	
	go func() {
		<-i.shutdown
		listener.Close()
	}()
	
	for {
		conn, err := listener.Accept()
		if err != nil {
			select {
			case <-i.shutdown:
				return
			default:
				log.Printf("TCP accept error: %v", err)
				continue
			}
		}
		
		go i.handleTCPConnection(conn)
	}
}

// handleTCPConnection processes a TCP connection
func (i *Ingestor) handleTCPConnection(conn net.Conn) {
	defer conn.Close()
	
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		var entry LogEntry
		if err := json.Unmarshal(scanner.Bytes(), &entry); err != nil {
			log.Printf("TCP JSON parse error: %v", err)
			continue
		}
		
		select {
		case i.logChan <- entry:
		case <-i.shutdown:
			return
		}
	}
	
	if err := scanner.Err(); err != nil {
		log.Printf("TCP scanner error: %v", err)
	}
}

// Stop gracefully shuts down the ingestor
func (i *Ingestor) Stop() {
	close(i.shutdown)
	i.wg.Wait()
	log.Println("Ingestor stopped")
}
