# Argos - Real-time Log Anomaly Detector

High-performance real-time log anomaly detector built in Go. Features concurrent ingestion, worker-pool parsing, and probabilistic threat flagging via custom Bloom Filters. Designed for massive throughput and O(1) memory efficiency.

## Architecture

```
[Log Generator (Python)] --(HTTP/TCP Traffic)--> [ 1. Ingestor Service (Go) ]
                                                          |
                                                  (Buffered Channel)
                                                          |
                                                  [ 2. Parser Workers (Go) ]
                                                          |
                                                   (Struct Channel)
                                                          |
                                                  [ 3. Analyzer Engine (Go) ]
                                                    /      |        \
                                              [Rules] [Bloom Filter] [Window Counter]
                                                          |
                                                  [ 4. Alerter (Go) ]
                                                          |
                                                   (Output/JSON)
```

## Features

- **Concurrent Ingestion**: HTTP and TCP endpoints for high-throughput log collection
- **Worker Pool Parsing**: Configurable number of concurrent parser workers
- **Bloom Filter**: Probabilistic data structure for O(1) pattern detection
- **Time Window Counting**: Track anomaly frequencies within sliding time windows
- **Rule-Based Detection**: Extensible rule system for custom anomaly detection
- **JSON Output**: Structured alert output to file and console

## Components

### 1. Ingestor Service
- Receives logs via HTTP POST (`/logs` endpoint) or TCP stream
- Non-blocking buffered channels for high throughput
- Graceful shutdown with connection draining

### 2. Parser Workers
- Worker pool pattern with configurable concurrency
- Extracts structured data from raw logs:
  - IP addresses
  - Error codes
  - Keywords
- Regular expression-based field extraction

### 3. Analyzer Engine
- **Rules Engine**: Detects critical errors, suspicious keywords, high error rates
- **Bloom Filter**: Probabilistic duplicate/pattern detection (100K capacity, 3 hash functions)
- **Window Counter**: Tracks anomaly frequency per source within 1-minute windows

### 4. Alerter
- JSON-formatted alert output
- Console and file logging
- Alert metadata includes pattern recognition and frequency counts

## Installation

### Prerequisites
- Go 1.16 or later
- Python 3.7+ (for log generator)

### Build

```bash
go build -o argos
```

## Usage

### Start Argos

```bash
./argos
```

The system will start with:
- HTTP endpoint: `http://localhost:8080/logs`
- TCP endpoint: `localhost:9090`
- Alert output: `alerts.json`

### Generate Test Logs

Install Python dependencies:
```bash
pip install requests
```

Run the log generator:
```bash
# HTTP mode (default)
python generator.py

# TCP mode
python generator.py --mode tcp

# Custom rate (10 logs/second)
python generator.py --rate 10

# Generate specific number of logs
python generator.py --count 1000

# Disable anomalies
python generator.py --no-anomalies
```

### Send Custom Logs

#### HTTP
```bash
curl -X POST http://localhost:8080/logs \
  -H "Content-Type: application/json" \
  -d '{
    "timestamp": "2024-01-15T10:30:00Z",
    "level": "ERROR",
    "source": "web-server-01",
    "message": "Database connection failed"
  }'
```

#### TCP
```bash
echo '{"timestamp":"2024-01-15T10:30:00Z","level":"CRITICAL","source":"api-gateway","message":"Unauthorized access from 192.168.1.100"}' | nc localhost 9090
```

## Configuration

Edit `main.go` to customize:
- `ingestBufferSize`: Ingestion channel buffer (default: 1000)
- `parseBufferSize`: Parser channel buffer (default: 1000)
- `alertBufferSize`: Alert channel buffer (default: 100)
- `httpPort`: HTTP server port (default: 8080)
- `tcpPort`: TCP server port (default: 9090)
- `parserWorkers`: Number of parser workers (default: 4)
- `alertOutputFile`: Alert output file (default: alerts.json)

## Alert Rules

Current detection rules:
1. **Critical Error Level**: Detects CRITICAL/FATAL log levels (HIGH severity)
2. **Error Code 5xx**: Detects 5xx HTTP error codes (HIGH severity)
3. **Suspicious Keywords**: Detects attack, breach, unauthorized, exploit, malicious (MEDIUM severity)
4. **Error Rate Threshold**: Tracks ERROR level frequency (MEDIUM severity)

## Performance

- **Concurrency**: Leverages Go goroutines for parallel processing
- **Memory Efficiency**: Bloom filter provides O(1) space complexity for pattern storage
- **Throughput**: Buffered channels prevent backpressure under load
- **Non-blocking**: All components use select statements for graceful degradation

## Development

### Project Structure
```
argos/
├── main.go              # Application entry point
├── ingestor/            # HTTP/TCP log ingestion
│   └── ingestor.go
├── parser/              # Log parsing and field extraction
│   └── parser.go
├── analyzer/            # Anomaly detection engine
│   ├── analyzer.go
│   └── bloomfilter.go
├── alerter/             # Alert output handler
│   └── alerter.go
└── generator.py         # Python log generator
```

### Testing

Run the system:
```bash
# Terminal 1: Start Argos
./argos

# Terminal 2: Generate logs
python generator.py --rate 5
```

Monitor alerts in console and `alerts.json`.

## License

MIT
