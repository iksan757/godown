#!/usr/bin/env bash

#
echo "⚡ Building godown binary..."

#
OUTPUT_BINARY="godown"

#
AUTHOR_NAME="Iksan rumasoreng"

#
LDFLAGS="-s -w -X 'main.AppName=godown' -X 'main.Author=${AUTHOR_NAME}' -X 'main.Version=1.0.0'"

#
eval go build -ldflags \"$LDFLAGS\" -o $OUTPUT_BINARY main.go

if [ $? -eq 0 ]; then
    echo "✅ Build successful! Binary created: ./$OUTPUT_BINARY"
    chmod +x $OUTPUT_BINARY
else
    echo "❌ Build failed!"
    exit 1
fi
