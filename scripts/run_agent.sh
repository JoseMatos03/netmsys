#!/bin/bash

# Check if an ID argument is provided
if [ -z "$1" ]; then
  echo "Usage: $0 <ID>"
  exit 1
fi

# Variables
ID=$1
SERVER_IP="10.0.0.20"
UDP_PORT=8080
TCP_PORT=9090

# Run the agent
./build/agent "$ID" "$SERVER_IP" "$UDP_PORT" "$TCP_PORT"
