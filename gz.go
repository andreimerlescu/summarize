package main

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
)

// compress compresses a string using gzip and returns the compressed bytes
func compress(s []byte) ([]byte, error) {
	var buf bytes.Buffer
	gzWriter := gzip.NewWriter(&buf)
	_, err := gzWriter.Write(s)
	if err != nil {
		return nil, fmt.Errorf("failed to write to gzip writer: %w", err)
	}
	err = gzWriter.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to close gzip writer: %w", err)
	}
	return buf.Bytes(), nil
}

// decompress decompresses gzip compressed bytes back to a string
func decompress(compressed []byte) (string, error) {
	buf := bytes.NewReader(compressed)
	gzReader, err := gzip.NewReader(buf)
	if err != nil {
		return "", fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer func() {
		_ = gzReader.Close()
	}()
	decompressed, err := io.ReadAll(gzReader)
	if err != nil {
		return "", fmt.Errorf("failed to read from gzip reader: %w", err)
	}
	return string(decompressed), nil
}
