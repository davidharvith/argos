#!/bin/bash
# Demo script for Argos Log Anomaly Detection System

echo "=========================================="
echo "  ARGOS - Real-time Log Anomaly Detector"
echo "=========================================="
echo ""
echo "Starting Argos in background..."
./argos > /tmp/argos_demo.log 2>&1 &
ARGOS_PID=$!
echo "Argos started (PID: $ARGOS_PID)"
sleep 2

echo ""
echo "Generating 10 test logs with anomalies..."
echo ""
python3 generator.py --count 10 --rate 3

echo ""
echo "Waiting for processing..."
sleep 2

echo ""
echo "=========================================="
echo "  Sample Alerts Generated:"
echo "=========================================="
if [ -f alerts.json ]; then
    echo ""
    head -30 alerts.json
    echo ""
    echo "... (see alerts.json for complete output)"
else
    echo "No alerts generated yet"
fi

echo ""
echo "=========================================="
echo "  System Statistics:"
echo "=========================================="
echo "Total alerts: $(wc -l < alerts.json 2>/dev/null || echo 0) lines"
echo "Binary size: $(du -h argos | cut -f1)"
echo "Source code: 892 lines (698 Go + 194 Python)"
echo ""

echo "Stopping Argos..."
kill $ARGOS_PID 2>/dev/null
wait $ARGOS_PID 2>/dev/null

echo ""
echo "Demo complete!"
