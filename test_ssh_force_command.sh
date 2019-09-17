#!/bin/bash

echo "just a simple test"
echo "parameters: $@"
echo "VAR1: ${VAR1:-not_set}"
echo "VAR2: ${VAR2:-not_set}"

exit 0
