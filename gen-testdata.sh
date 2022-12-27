#!/usr/bin/env bash

for file in "./testdata"/*; do
	if ! echo "$file" | grep "_encoding.go$"; then
		go generate "$file"
		dst="${file%\.go}_encoding.go"
		mv ./testdata/main_encoding.go "$dst"
	fi
done
