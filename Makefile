# Directory where the binaries will be placed
BUILD_DIR = build

# Finds all *.go files
GO_FILES = $(shell find . -name '*.go')

# Generate binary names based on Go filenames (remove directories and .go extensions)
BINS = $(patsubst ./%.go,$(BUILD_DIR)/%,$(GO_FILES))

# Default target
all: fmtcheck importscheck $(BINS)

# Formatting check
fmtcheck:
	@scripts/gofmtcheck.sh

# Goimports check
importscheck:
	@scripts/goimportscheck.sh

# Rule to compile Go files into binaries
$(BUILD_DIR)/%: %.go
	@mkdir -p $(dir $@)
	go build -o $@ $<

# Clean target to remove binaries
clean:
	rm -rf $(BUILD_DIR)

# Probably not needed in this applications' current state, but we've decided to
# be better safe than sorry and disable parallelism altogether.
.NOTPARALLEL:

.PHONY: all clean fmtcheck importscheck
