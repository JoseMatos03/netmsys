#!/bin/bash

# Find all go files
unformatted_files=$(gofmt -l .)

# Check if there are any unformatted files
if [ -n "$unformatted_files" ]; then
    echo "The following files are not properly formatted:"
    echo "$unformatted_files"
    echo "Please run 'go fmt' on the above files."
    exit 1
fi

exit 0
