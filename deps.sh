#!/bin/bash

FILE=/tmp/velonetics_ce_deps.txt

go list -m -u all > "$FILE"

OUTPUT=$(grep -r "\[" "$FILE" | grep velonetics | sed 's/\[//g' | sed 's/\]//g' | awk '{print "go get", $1"@"$3 }')

if [ "$OUTPUT" != "" ]; then
	echo "$OUTPUT"
	echo "go mod tidy"
	exit 1
fi

echo "all deps up to date."
