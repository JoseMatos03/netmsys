# Directory where the binaries will be placed
BUILD_DIR = build
OUTPUT_DIR = outputs

# Paths to main Go files for each component
AGENT_MAIN = cmd/agent/agent.go
SERVER_MAIN = cmd/server/server.go

# Default target to build both server and agent binaries
all: fmtcheck importscheck build-server build-agent

# Formatting check
fmtcheck:
	@scripts/gofmtcheck.sh

# Goimports check
importscheck:
	@scripts/goimportscheck.sh

# Rule to build the server binary
build-server:
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/server $(SERVER_MAIN)

# Rule to build the agent binary
build-agent:
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/agent $(AGENT_MAIN)

# Clean target to remove binaries
clean:
	rm -rf $(BUILD_DIR)
	rm -rf $(OUTPUT_DIR)

# Disable parallelism to avoid issues
.NOTPARALLEL:

.PHONY: all clean fmtcheck importscheck build-server build-agent
