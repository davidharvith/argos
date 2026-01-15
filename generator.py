#!/usr/bin/env python3
"""
Log Generator for Argos
Generates synthetic log data and sends it to the Argos ingestor via HTTP or TCP
"""

import json
import random
import time
import argparse
from datetime import datetime
import requests
import socket

# Log levels with weights for realistic distribution
LOG_LEVELS = [
    ("INFO", 70),
    ("WARN", 20),
    ("ERROR", 8),
    ("CRITICAL", 1.5),
    ("FATAL", 0.5),
]

# Log sources
SOURCES = [
    "web-server-01",
    "web-server-02",
    "api-gateway",
    "database-primary",
    "database-replica",
    "auth-service",
    "cache-redis",
    "payment-processor",
]

# Message templates
MESSAGES = [
    "Request processed successfully",
    "Database query executed in {time}ms",
    "User {user_id} logged in from {ip}",
    "Cache hit for key {key}",
    "API call to {endpoint} completed",
    "Connection timeout to {ip}",
    "ERROR: Failed to connect to database",
    "CRITICAL: Out of memory error",
    "Unauthorized access attempt from {ip}",
    "Payment transaction {transaction_id} completed",
    "500 Internal Server Error on {endpoint}",
    "503 Service Unavailable",
    "Authentication failed for user {user_id}",
    "Potential security breach detected from {ip}",
    "Suspicious activity: Multiple failed login attempts from {ip}",
    "System overload: Request queue at 95% capacity",
    "Database connection pool exhausted",
    "Failed to process payment for transaction {transaction_id}",
]

# Suspicious/anomaly messages (lower probability)
ANOMALY_MESSAGES = [
    "CRITICAL: Unauthorized access attempt detected from {ip}",
    "FATAL: System crash imminent - memory exhausted",
    "Security breach: SQL injection attack from {ip}",
    "Malicious payload detected in request from {ip}",
    "Exploit attempt: Buffer overflow detected",
    "Unauthorized API key usage from {ip}",
    "CRITICAL: Data corruption detected in table users",
]


def generate_log_entry(include_anomalies=True):
    """Generate a single log entry"""
    # Decide if this should be an anomaly
    is_anomaly = include_anomalies and random.random() < 0.05
    
    # Select log level
    level = random.choices(
        [level for level, _ in LOG_LEVELS],
        weights=[weight for _, weight in LOG_LEVELS]
    )[0]
    
    # Select source
    source = random.choice(SOURCES)
    
    # Generate message
    if is_anomaly:
        message_template = random.choice(ANOMALY_MESSAGES)
        if level in ["INFO", "WARN"]:
            level = random.choice(["CRITICAL", "FATAL", "ERROR"])
    else:
        message_template = random.choice(MESSAGES)
    
    # Fill in template variables
    message = message_template.format(
        time=random.randint(10, 5000),
        user_id=f"user_{random.randint(1000, 9999)}",
        ip=f"{random.randint(1, 255)}.{random.randint(1, 255)}.{random.randint(1, 255)}.{random.randint(1, 255)}",
        key=f"cache_key_{random.randint(100, 999)}",
        endpoint=random.choice(["/api/users", "/api/orders", "/api/products", "/api/auth"]),
        transaction_id=f"txn_{random.randint(10000, 99999)}",
    )
    
    return {
        "timestamp": datetime.utcnow().isoformat() + "Z",
        "level": level,
        "source": source,
        "message": message,
    }


def send_http(log_entry, url):
    """Send log entry via HTTP"""
    try:
        response = requests.post(url, json=log_entry, timeout=2)
        return response.status_code == 200
    except Exception as e:
        print(f"HTTP error: {e}")
        return False


def send_tcp(log_entry, host, port):
    """Send log entry via TCP"""
    try:
        with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as sock:
            sock.settimeout(2)
            sock.connect((host, port))
            sock.sendall(json.dumps(log_entry).encode() + b'\n')
        return True
    except Exception as e:
        print(f"TCP error: {e}")
        return False


def main():
    parser = argparse.ArgumentParser(description="Generate synthetic logs for Argos")
    parser.add_argument("--mode", choices=["http", "tcp"], default="http",
                        help="Transport mode (default: http)")
    parser.add_argument("--host", default="localhost",
                        help="Argos host (default: localhost)")
    parser.add_argument("--http-port", type=int, default=8080,
                        help="HTTP port (default: 8080)")
    parser.add_argument("--tcp-port", type=int, default=9090,
                        help="TCP port (default: 9090)")
    parser.add_argument("--rate", type=float, default=2.0,
                        help="Logs per second (default: 2.0)")
    parser.add_argument("--count", type=int, default=0,
                        help="Number of logs to generate (0 = infinite)")
    parser.add_argument("--no-anomalies", action="store_true",
                        help="Disable anomaly generation")
    
    args = parser.parse_args()
    
    url = f"http://{args.host}:{args.http_port}/logs"
    delay = 1.0 / args.rate
    
    print(f"Starting log generator...")
    print(f"Mode: {args.mode}")
    print(f"Rate: {args.rate} logs/second")
    print(f"Anomalies: {'disabled' if args.no_anomalies else 'enabled'}")
    
    if args.mode == "http":
        print(f"Target: {url}")
    else:
        print(f"Target: {args.host}:{args.tcp_port}")
    
    print("\nPress Ctrl+C to stop\n")
    
    count = 0
    try:
        while args.count == 0 or count < args.count:
            log_entry = generate_log_entry(include_anomalies=not args.no_anomalies)
            
            success = False
            if args.mode == "http":
                success = send_http(log_entry, url)
            else:
                success = send_tcp(log_entry, args.host, args.tcp_port)
            
            if success:
                count += 1
                status = "✓"
                if log_entry["level"] in ["CRITICAL", "FATAL"]:
                    status = "🚨"
                print(f"{status} [{count}] {log_entry['level']:8s} {log_entry['source']:20s} {log_entry['message'][:60]}")
            else:
                print(f"✗ Failed to send log")
            
            time.sleep(delay)
    
    except KeyboardInterrupt:
        print(f"\n\nStopped. Generated {count} logs.")


if __name__ == "__main__":
    main()
