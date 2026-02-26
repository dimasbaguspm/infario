package engine

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// FileEngine handles generic filesystem CRUD operations for binary assets.
// The consumer is responsible for determining the directory structure.
type FileEngine struct {
	BaseDir string
}

// NewFileEngine creates a new FileEngine with the given base directory.
func NewFileEngine(baseDir string) *FileEngine {
	return &FileEngine{BaseDir: baseDir}
}

// Store extracts archive content (zip or tar.gz) into the given path with content-addressable deduplication.
// path should be relative to BaseDir (e.g., "deployments/project-id/hash/filename.tar.gz").
// Automatically detects format from filename extension.
// Returns true if extraction happened, false if path already exists (deduplication).
// Uses atomic rename to ensure consistency on failure.
func (e *FileEngine) Store(ctx context.Context, path string, archiveData io.Reader) (bool, error) {
	finalPath := filepath.Join(e.BaseDir, path)
	tmpPath := filepath.Join(e.BaseDir, path+"_tmp")

	// 1. Content-Addressable Check (Deduplication)
	if _, err := os.Stat(finalPath); err == nil {
		return false, nil
	}

	// 2. Prepare Directories
	if err := os.MkdirAll(filepath.Dir(finalPath), 0755); err != nil {
		return false, fmt.Errorf("failed to create directories: %w", err)
	}

	// Cleanup any failed previous attempts
	_ = os.RemoveAll(tmpPath)

	// 3. Extract to Temporary Folder (detect format from filename)
	destDir := tmpPath
	if strings.HasSuffix(path, ".tar.gz") || strings.HasSuffix(path, ".tgz") {
		if err := e.extractTarGz(archiveData, destDir); err != nil {
			_ = os.RemoveAll(tmpPath)
			return false, fmt.Errorf("tar.gz extraction failed: %w", err)
		}
	} else {
		if err := e.unzip(archiveData, destDir); err != nil {
			_ = os.RemoveAll(tmpPath)
			return false, fmt.Errorf("zip extraction failed: %w", err)
		}
	}

	// 4. Atomic Rename (The "Commit")
	if err := os.Rename(tmpPath, finalPath); err != nil {
		_ = os.RemoveAll(tmpPath)
		return false, fmt.Errorf("failed to finalize storage: %w", err)
	}

	return true, nil
}

// Remove cleans up assets from the disk at the given path.
// path should be relative to BaseDir (same format as Store).
func (e *FileEngine) Remove(ctx context.Context, path string) error {
	fullPath := filepath.Join(e.BaseDir, path)
	return os.RemoveAll(fullPath)
}

// extractTarGz safely extracts tar.gz content to destination directory with ZIP SLIP protection.
func (e *FileEngine) extractTarGz(src io.Reader, dest string) error {
	gzr, err := gzip.NewReader(src)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read tar entry: %w", err)
		}

		// ZIP SLIP protection: ensure file path is within destination
		path := filepath.Join(dest, header.Name)
		if !strings.HasPrefix(path, filepath.Clean(dest)+string(os.PathSeparator)) {
			return fmt.Errorf("illegal file path: %s", header.Name)
		}

		if header.Typeflag == tar.TypeDir {
			if err := os.MkdirAll(path, 0755); err != nil {
				return fmt.Errorf("failed to create directory: %w", err)
			}
		} else {
			if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
				return fmt.Errorf("failed to create parent directory: %w", err)
			}

			file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
			if err != nil {
				return fmt.Errorf("failed to create file: %w", err)
			}

			if _, err := io.Copy(file, tr); err != nil {
				file.Close()
				return fmt.Errorf("failed to write file: %w", err)
			}
			file.Close()
		}
	}

	return nil
}

// unzip safely extracts zip content to destination directory.
// Buffers the stream to a temporary file since zip.NewReader requires ReaderAt.
func (e *FileEngine) unzip(src io.Reader, dest string) error {
	tmpZip, err := os.CreateTemp("", "infario-upload-*.zip")
	if err != nil {
		return err
	}
	defer os.Remove(tmpZip.Name())
	defer tmpZip.Close()

	size, err := io.Copy(tmpZip, src)
	if err != nil {
		return err
	}

	reader, err := zip.NewReader(tmpZip, size)
	if err != nil {
		return err
	}

	for _, f := range reader.File {
		if err := e.extractFile(f, dest); err != nil {
			return err
		}
	}

	return nil
}

// extractFile safely extracts a single file from a zip with ZIP SLIP protection.
func (e *FileEngine) extractFile(f *zip.File, dest string) error {

	path := filepath.Join(dest, f.Name)
	if !strings.HasPrefix(path, filepath.Clean(dest)+string(os.PathSeparator)) {
		return fmt.Errorf("illegal file path: %s", f.Name)
	}

	if f.FileInfo().IsDir() {
		return os.MkdirAll(path, f.Mode())
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	dstFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
	if err != nil {
		return err
	}
	defer dstFile.Close()

	srcFile, err := f.Open()
	if err != nil {
		return err
	}
	defer srcFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}
