package transfer

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"sort"
)

// HashPrefix is the prefix for SHA-256 hashes.
const HashPrefix = "sha256:"

// HashFile computes the SHA-256 hash of a file.
// Returns the hash in format "sha256:<hex>".
func HashFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, f); err != nil {
		return "", err
	}

	return HashPrefix + hex.EncodeToString(hasher.Sum(nil)), nil
}

// HashBytes computes the SHA-256 hash of a byte slice.
// Returns the hash in format "sha256:<hex>".
func HashBytes(data []byte) string {
	hasher := sha256.New()
	hasher.Write(data)
	return HashPrefix + hex.EncodeToString(hasher.Sum(nil))
}

// ComputeContentHash computes the overall content hash from a list of file hashes.
// Files are sorted by path for deterministic ordering before hash computation.
// Returns the hash in format "sha256:<hex>".
func ComputeContentHash(files []FileInfo) string {
	if len(files) == 0 {
		return ""
	}

	// Sort files by path for deterministic ordering
	sorted := make([]FileInfo, len(files))
	copy(sorted, files)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Path < sorted[j].Path
	})

	// Concatenate hashes and compute final hash
	hasher := sha256.New()
	for _, file := range sorted {
		hasher.Write([]byte(file.Hash))
	}

	return HashPrefix + hex.EncodeToString(hasher.Sum(nil))
}
