package analyzer

import (
	"log"
	"sync"
	"time"

	"github.com/davidharvith/argos/parser"
)

// Alert represents a detected anomaly
type Alert struct {
	Timestamp string                 `json:"timestamp"`
	Severity  string                 `json:"severity"`
	Reason    string                 `json:"reason"`
	Log       parser.ParsedLog       `json:"log"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// Rule defines an anomaly detection rule
type Rule struct {
	Name      string
	Check     func(parser.ParsedLog) bool
	Severity  string
}

// Analyzer processes parsed logs and detects anomalies
type Analyzer struct {
	inputChan    <-chan parser.ParsedLog
	alertChan    chan<- Alert
	rules        []Rule
	bloomFilter  *BloomFilter
	windowCount  map[string]int
	windowMutex  sync.RWMutex
	windowSize   time.Duration
	shutdown     chan struct{}
	wg           sync.WaitGroup
}

// NewAnalyzer creates a new Analyzer instance
func NewAnalyzer(inputChan <-chan parser.ParsedLog, alertChan chan<- Alert) *Analyzer {
	a := &Analyzer{
		inputChan:   inputChan,
		alertChan:   alertChan,
		bloomFilter: NewBloomFilter(100000, 3),
		windowCount: make(map[string]int),
		windowSize:  time.Minute,
		shutdown:    make(chan struct{}),
	}
	
	// Initialize default rules
	a.initializeRules()
	
	return a
}

// initializeRules sets up the default anomaly detection rules
func (a *Analyzer) initializeRules() {
	a.rules = []Rule{
		{
			Name: "Critical Error Level",
			Check: func(log parser.ParsedLog) bool {
				return log.Level == "CRITICAL" || log.Level == "FATAL"
			},
			Severity: "HIGH",
		},
		{
			Name: "Error Code 5xx",
			Check: func(log parser.ParsedLog) bool {
				return len(log.ErrorCode) > 0 && log.ErrorCode[0] == '5'
			},
			Severity: "HIGH",
		},
		{
			Name: "Suspicious Keywords",
			Check: func(log parser.ParsedLog) bool {
				suspiciousWords := []string{"attack", "breach", "unauthorized", "exploit", "malicious"}
				for _, kw := range log.Keywords {
					for _, sw := range suspiciousWords {
						if kw == sw {
							return true
						}
					}
				}
				return false
			},
			Severity: "MEDIUM",
		},
		{
			Name: "Error Rate Threshold",
			Check: func(log parser.ParsedLog) bool {
				return log.Level == "ERROR"
			},
			Severity: "MEDIUM",
		},
	}
}

// Start begins the analyzer
func (a *Analyzer) Start() {
	a.wg.Add(2)
	go a.analyze()
	go a.cleanupWindow()
	log.Println("Analyzer started")
}

// analyze processes logs and detects anomalies
func (a *Analyzer) analyze() {
	defer a.wg.Done()
	
	for {
		select {
		case logEntry, ok := <-a.inputChan:
			if !ok {
				return
			}
			a.processLog(logEntry)
		case <-a.shutdown:
			return
		}
	}
}

// processLog checks a log against all rules and generates alerts
func (a *Analyzer) processLog(logEntry parser.ParsedLog) {
	for _, rule := range a.rules {
		if rule.Check(logEntry) {
			// Check if we've seen similar patterns recently
			bloomKey := rule.Name + ":" + logEntry.Source
			isKnownPattern := a.bloomFilter.Contains(bloomKey)
			a.bloomFilter.Add(bloomKey)
			
			// Track frequency in time window
			a.windowMutex.Lock()
			countKey := rule.Name + ":" + logEntry.Source
			a.windowCount[countKey]++
			count := a.windowCount[countKey]
			a.windowMutex.Unlock()
			
			// Create alert
			alert := Alert{
				Timestamp: time.Now().Format(time.RFC3339),
				Severity:  rule.Severity,
				Reason:    rule.Name,
				Log:       logEntry,
				Metadata: map[string]interface{}{
					"is_known_pattern": isKnownPattern,
					"count_in_window":  count,
					"rule_name":        rule.Name,
				},
			}
			
			select {
			case a.alertChan <- alert:
			case <-a.shutdown:
				return
			}
		}
	}
}

// cleanupWindow periodically resets the time window counters
func (a *Analyzer) cleanupWindow() {
	defer a.wg.Done()
	
	ticker := time.NewTicker(a.windowSize)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			a.windowMutex.Lock()
			a.windowCount = make(map[string]int)
			a.windowMutex.Unlock()
			log.Println("Window counters reset")
		case <-a.shutdown:
			return
		}
	}
}

// Stop gracefully shuts down the analyzer
func (a *Analyzer) Stop() {
	close(a.shutdown)
	a.wg.Wait()
	log.Println("Analyzer stopped")
}
