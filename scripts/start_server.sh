#!/bin/bash

# Ports for the server
UDP_PORT=8080
TCP_PORT=9090

# Commands to send to the server
cat <<EOF | ./build/server $UDP_PORT $TCP_PORT
load ./configs/tasks/Agent1Task.json
load ./configs/tasks/Agent2Task.json
load ./configs/tasks/Agent3Task.json
EOF