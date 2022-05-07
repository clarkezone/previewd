#!/usr/bin/env bash

# Use this by
# 1. creating env file
# 2. source setenv.sh

# Show env vars
grep -v '^#' ~/.previewd_test.env

# Export env vars
export "$(grep -v '^#' ~/.previewd_test.env | xargs)"
