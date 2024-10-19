#!/bin/bash

# Find all go files that don't comply with goimports
unformatted_imports=$(goimports -l .)

# Check if there are any files that need goimports fixing
if [ -n "$unformatted_imports" ]; then
    echo "The following files do not comply with goimports:"
    echo "$unformatted_imports"
    echo "Please run 'goimports' on the above files."
    exit 1
fi

exit 0
