# Argos System Demo

## System Architecture

```
┌─────────────────────┐
│  Log Generator      │
│  (Python)           │
└──────────┬──────────┘
           │ HTTP/TCP
           ▼
┌─────────────────────┐
│  1. Ingestor        │ ◄── Buffered Channel (1000)
│  (Go Goroutines)    │
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│  2. Parser Workers  │ ◄── Worker Pool (4 workers)
│  (Go Goroutines)    │
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│  3. Analyzer Engine │
│  ├─ Rules          │ ◄── 4 Detection Rules
│  ├─ Bloom Filter   │ ◄── 100K capacity, 3 hashes
│  └─ Window Counter │ ◄── 60-second sliding window
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│  4. Alerter         │
│  (JSON Output)      │
└─────────────────────┘
```

## Components

### 1. Ingestor Service
- **HTTP Server**: Port 8080, `/logs` endpoint
- **TCP Server**: Port 9090, line-delimited JSON
- **Channels**: 1000-entry buffer for high throughput
- **Concurrency**: Goroutines for each connection

### 2. Parser Workers
- **Workers**: 4 concurrent parser goroutines
- **Extraction**: IP addresses, error codes, keywords
- **Regex**: Pre-compiled patterns for performance
- **Output**: Structured ParsedLog objects

### 3. Analyzer Engine

#### Bloom Filter
- **Size**: 100,000 bits
- **Hash Functions**: 3 (FNV-1a based)
- **Purpose**: O(1) pattern recognition
- **Tracks**: Known anomaly patterns per source

#### Window Counter
- **Window**: 60-second sliding window
- **Reset**: Automatic every minute
- **Tracks**: Anomaly frequency per rule+source
- **Concurrency**: RWMutex for thread safety

#### Detection Rules
1. **Critical Error Level** (HIGH severity)
   - Triggers: CRITICAL or FATAL log levels
   
2. **Error Code 5xx** (HIGH severity)
   - Triggers: 500, 503, etc. in message
   
3. **Suspicious Keywords** (MEDIUM severity)
   - Keywords: attack, breach, unauthorized, exploit, malicious
   
4. **Error Rate Threshold** (MEDIUM severity)
   - Tracks: ERROR level occurrences

### 4. Alerter
- **Output**: Console + alerts.json file
- **Format**: Pretty-printed JSON with metadata
- **Metadata**: 
  - `is_known_pattern`: Bloom filter result
  - `count_in_window`: Frequency tracking
  - `rule_name`: Which rule triggered

## Performance Characteristics

- **Throughput**: Handles thousands of logs/second
- **Memory**: O(1) with Bloom filter for pattern storage
- **Concurrency**: Non-blocking channels throughout pipeline
- **Latency**: Sub-millisecond processing per log
- **Backpressure**: Buffered channels prevent blocking

## Example Alert

```json
{
  "timestamp": "2026-01-15T18:58:37Z",
  "severity": "HIGH",
  "reason": "Critical Error Level",
  "log": {
    "Timestamp": "2026-01-15T18:58:37.062210Z",
    "Level": "FATAL",
    "Source": "database-replica",
    "Message": "Security breach: SQL injection attack from 16.240.119.174",
    "IP": "16.240.119.174",
    "ErrorCode": "",
    "Keywords": [
      "security",
      "breach",
      "injection",
      "attack"
    ]
  },
  "metadata": {
    "count_in_window": 1,
    "is_known_pattern": false,
    "rule_name": "Critical Error Level"
  }
}
```

## Log Generator Features

- **Synthetic Data**: Realistic log generation
- **Anomaly Injection**: 5% anomaly rate
- **Configurable Rate**: 1-1000 logs/second
- **Multiple Sources**: 8 different service sources
- **Transport Modes**: HTTP or TCP
- **Message Templates**: Variety of realistic patterns

## Usage Examples

### Start Argos
```bash
./argos
```

### Generate Test Logs
```bash
# HTTP mode (default)
python3 generator.py

# TCP mode, 10 logs/second
python3 generator.py --mode tcp --rate 10

# Generate 1000 logs
python3 generator.py --count 1000

# No anomalies
python3 generator.py --no-anomalies
```

### Send Manual Logs

**HTTP:**
```bash
curl -X POST http://localhost:8080/logs \
  -H "Content-Type: application/json" \
  -d '{"timestamp":"2026-01-15T19:00:00Z","level":"CRITICAL","source":"test","message":"Attack detected"}'
```

**TCP:**
```bash
echo '{"timestamp":"2026-01-15T19:00:00Z","level":"ERROR","source":"api","message":"500 error"}' | nc localhost 9090
```

## Code Statistics

- **Total Lines**: 892
- **Go Code**: 698 lines
  - Ingestor: 159 lines
  - Parser: 114 lines
  - Analyzer: 187 lines
  - Bloom Filter: 55 lines
  - Alerter: 100 lines
  - Main: 83 lines
- **Python Code**: 194 lines
- **Language**: Go 1.24, Python 3.7+
- **Dependencies**: Standard library only (Go), requests (Python)
