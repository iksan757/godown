#!/bin/bash

# Banner
echo "⚡ Building godown binary..."

# Nama output binary
OUTPUT_BINARY="godown"

# Target nama pembuat (sesuai string asli di main.go)
AUTHOR_NAME="Iksan  rumasoreng"

# Flag kompilasi:
# -s -w : Menghapus informasi debugging & symbol table (memperkecil ukuran binary)
# -X    : Menginjeksi metadata aplikasi saat kompilasi
LDFLAGS="-s -w -X 'main.AppName=godown' -X 'main.Author=${AUTHOR_NAME}' -X 'main.Version=1.0.0'"

# Proses kompilasi Go
eval go build -ldflags \"$LDFLAGS\" -o $OUTPUT_BINARY main.go

if [ $? -eq 0 ]; then
    echo "✅ Build successful! Binary created: ./$OUTPUT_BINARY"
    chmod +x $OUTPUT_BINARY
else
    echo "❌ Build failed!"
    exit 1
fi
