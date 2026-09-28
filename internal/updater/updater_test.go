package updater

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"os"
	"path/filepath"
	"testing"
)

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		v1       string
		v2       string
		expected int
	}{
		{"v1.0.3", "v1.0.4", -1},
		{"v1.0.4", "v1.0.3", 1},
		{"v1.0.3", "v1.0.3", 0},
		{"v1.0.3-dirty", "v1.0.3", 0},
		{"v1.0.3+build", "v1.0.3", 0},
		{"1.0.3", "v1.0.3", 0},
		{"v1.1.0", "v1.0.9", 1},
		{"v2.0.0", "v1.99.99", 1},
		{"dev", "v1.0.0", -1},
		{"v1.0.0", "dev", 1},
		{"dev", "dev", 0},
	}

	for _, tt := range tests {
		got := CompareVersions(tt.v1, tt.v2)
		if got != tt.expected {
			t.Errorf("CompareVersions(%q, %q) = %d, expected %d", tt.v1, tt.v2, got, tt.expected)
		}
	}
}

func TestFindAsset(t *testing.T) {
	assets := []Asset{
		{Name: "online-mermaid-v1.0.4-linux-amd64.tar.gz", BrowserDownloadURL: "https://example.com/linux-amd64.tar.gz"},
		{Name: "online-mermaid-v1.0.4-linux-amd64.tar.gz.sha256", BrowserDownloadURL: "https://example.com/linux-amd64.tar.gz.sha256"},
		{Name: "online-mermaid-v1.0.4-windows-amd64.zip", BrowserDownloadURL: "https://example.com/windows-amd64.zip"},
		{Name: "online-mermaid-v1.0.4-windows-amd64.zip.sha256", BrowserDownloadURL: "https://example.com/windows-amd64.zip.sha256"},
		{Name: "online-mermaid-v1.0.4-darwin-arm64.tar.gz", BrowserDownloadURL: "https://example.com/darwin-arm64.tar.gz"},
	}

	// 1. Linux amd64
	archive, sha := FindAsset(assets, "linux", "amd64", "v1.0.4")
	if archive == nil || archive.Name != "online-mermaid-v1.0.4-linux-amd64.tar.gz" {
		t.Fatalf("expected linux archive asset, got %+v", archive)
	}
	if sha == nil || sha.Name != "online-mermaid-v1.0.4-linux-amd64.tar.gz.sha256" {
		t.Fatalf("expected linux sha asset, got %+v", sha)
	}

	// 2. Windows amd64
	winArchive, winSha := FindAsset(assets, "windows", "amd64", "v1.0.4")
	if winArchive == nil || winArchive.Name != "online-mermaid-v1.0.4-windows-amd64.zip" {
		t.Fatalf("expected windows archive asset, got %+v", winArchive)
	}
	if winSha == nil || winSha.Name != "online-mermaid-v1.0.4-windows-amd64.zip.sha256" {
		t.Fatalf("expected windows sha asset, got %+v", winSha)
	}

	// 3. Darwin arm64
	macArchive, macSha := FindAsset(assets, "darwin", "arm64", "v1.0.4")
	if macArchive == nil || macArchive.Name != "online-mermaid-v1.0.4-darwin-arm64.tar.gz" {
		t.Fatalf("expected mac archive asset, got %+v", macArchive)
	}
	if macSha != nil {
		t.Fatalf("expected nil mac sha, got %+v", macSha)
	}
}

func TestExtractFromTarGz(t *testing.T) {
	tempDir := t.TempDir()
	archivePath := filepath.Join(tempDir, "test.tar.gz")

	content := []byte("#!/bin/sh\necho 'new-version'\n")

	// Create test tar.gz
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)

	header := &tar.Header{
		Name: "online-mermaid",
		Mode: 0755,
		Size: int64(len(content)),
	}
	if err := tw.WriteHeader(header); err != nil {
		t.Fatalf("write tar header: %v", err)
	}
	if _, err := tw.Write(content); err != nil {
		t.Fatalf("write tar content: %v", err)
	}
	tw.Close()
	gw.Close()

	if err := os.WriteFile(archivePath, buf.Bytes(), 0644); err != nil {
		t.Fatalf("write archive file: %v", err)
	}

	var extracted bytes.Buffer
	if err := extractFromTarGz(archivePath, &extracted); err != nil {
		t.Fatalf("extractFromTarGz failed: %v", err)
	}

	if !bytes.Equal(extracted.Bytes(), content) {
		t.Fatalf("extracted content mismatch: got %q, expected %q", extracted.String(), string(content))
	}
}
