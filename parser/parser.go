package parser

import (
	"log"
	"regexp"
	"strings"
	"sync"

	"github.com/davidharvith/argos/ingestor"
)

// ParsedLog represents a parsed log entry with extracted fields
type ParsedLog struct {
	Timestamp string
	Level     string
	Source    string
	Message   string
	IP        string
	ErrorCode string
	Keywords  []string
}

// Parser processes raw log entries and extracts structured data
type Parser struct {
	inputChan  <-chan ingestor.LogEntry
	outputChan chan<- ParsedLog
	workers    int
	wg         sync.WaitGroup
	shutdown   chan struct{}
	ipRegex    *regexp.Regexp
	errorRegex *regexp.Regexp
}

// NewParser creates a new Parser instance
func NewParser(inputChan <-chan ingestor.LogEntry, outputChan chan<- ParsedLog, workers int) *Parser {
	return &Parser{
		inputChan:  inputChan,
		outputChan: outputChan,
		workers:    workers,
		shutdown:   make(chan struct{}),
		ipRegex:    regexp.MustCompile(`\b(?:\d{1,3}\.){3}\d{1,3}\b`),
		errorRegex: regexp.MustCompile(`\b(?:ERROR|FATAL|CRITICAL|[45]\d{2})\b`),
	}
}

// Start begins the parser workers
func (p *Parser) Start() {
	for i := 0; i < p.workers; i++ {
		p.wg.Add(1)
		go p.worker(i)
	}
	log.Printf("Started %d parser workers", p.workers)
}

// worker processes logs from the input channel
func (p *Parser) worker(id int) {
	defer p.wg.Done()
	
	for {
		select {
		case entry, ok := <-p.inputChan:
			if !ok {
				return
			}
			parsed := p.parse(entry)
			select {
			case p.outputChan <- parsed:
			case <-p.shutdown:
				return
			}
		case <-p.shutdown:
			return
		}
	}
}

// parse extracts structured data from a log entry
func (p *Parser) parse(entry ingestor.LogEntry) ParsedLog {
	parsed := ParsedLog{
		Timestamp: entry.Timestamp,
		Level:     entry.Level,
		Source:    entry.Source,
		Message:   entry.Message,
		Keywords:  []string{},
	}
	
	// Extract IP address
	if ip := p.ipRegex.FindString(entry.Message); ip != "" {
		parsed.IP = ip
	}
	
	// Extract error codes
	if errCode := p.errorRegex.FindString(entry.Message); errCode != "" {
		parsed.ErrorCode = errCode
	}
	
	// Extract keywords (simple tokenization)
	words := strings.Fields(entry.Message)
	for _, word := range words {
		word = strings.ToLower(strings.Trim(word, ".,;:!?"))
		if len(word) > 3 {
			parsed.Keywords = append(parsed.Keywords, word)
		}
	}
	
	return parsed
}

// Stop gracefully shuts down the parser
func (p *Parser) Stop() {
	close(p.shutdown)
	p.wg.Wait()
	log.Println("Parser stopped")
}
