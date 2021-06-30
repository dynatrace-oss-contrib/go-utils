#!/bin/bash

command -v go-licence-detector >/dev/null 2>&1 || { echo >&2 "This script requires 'go-licence-detector' to be installed. Please install it via 'go get go.elastic.co/go-licence-detector'"; exit 1; }

SCRIPT_FULL_PATH=$(dirname "$0")
cd "$SCRIPT_FULL_PATH" || return

echo "🔍 Analyzing dependencies in module go-utils"

( cd .. || return ; go list -m -json all | go-licence-detector -depsTemplate=.licences/templates/dependencies.csv.tmpl -depsOut=.licences/dependencies.txt )

echo "👍 done. written results to ./dependencies.txt"