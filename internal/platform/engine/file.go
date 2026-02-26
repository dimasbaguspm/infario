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

// FileEngine handles archive extraction and filesystem operations.
type FileEngine struct {
	BaseDir string
}

// NewFileEngine creates a new FileEngine with the given base directory.
func NewFileEngine(baseDir string) *FileEngine {
	return &FileEngine{BaseDir: baseDir}
}

// Extract extracts archive content (zip or tar.gz) into the destination directory.
// If the archive contains a single root directory, its contents are extracted directly.
// Otherwise, all archive contents are extracted as-is.
// destPath should be relative to BaseDir (e.g., "deployments/project-id/deployment-id").
// Automatically detects format from filename extension.
func (e *FileEngine) Extract(ctx context.Context, destPath string, archiveData io.Reader, filename string) error {
	fullDestPath := filepath.Join(e.BaseDir, destPath)
	tempPath := filepath.Join(e.BaseDir, destPath+"_extract_tmp")

	// Create temporary directory for extraction
	if err := os.MkdirAll(tempPath, 0755); err != nil {
		return fmt.Errorf("failed to create temporary directory: %w", err)
	}
	defer os.RemoveAll(tempPath)

	// Detect format and extract to temporary location
	if strings.HasSuffix(filename, ".tar.gz") || strings.HasSuffix(filename, ".tgz") {
		if err := e.extractTarGz(archiveData, tempPath); err != nil {
			return fmt.Errorf("tar.gz extraction failed: %w", err)
		}
	} else if strings.HasSuffix(filename, ".zip") {
		if err := e.unzip(archiveData, tempPath); err != nil {
			return fmt.Errorf("zip extraction failed: %w", err)
		}
	} else {
		return fmt.Errorf("unsupported archive format: %s", filename)
	}

	// Check if extracted content has a single root directory
	entries, err := os.ReadDir(tempPath)
	if err != nil {
		return fmt.Errorf("failed to read extracted contents: %w", err)
	}

	// If single root directory, move its contents up one level
	if len(entries) == 1 && entries[0].IsDir() {
		singleDirPath := filepath.Join(tempPath, entries[0].Name())
		subEntries, err := os.ReadDir(singleDirPath)
		if err == nil && len(subEntries) > 0 {
			// Create destination and move contents from subdirectory
			if err := os.MkdirAll(fullDestPath, 0755); err != nil {
				return fmt.Errorf("failed to create destination directory: %w", err)
			}
			for _, entry := range subEntries {
				src := filepath.Join(singleDirPath, entry.Name())
				dst := filepath.Join(fullDestPath, entry.Name())
				if err := os.Rename(src, dst); err != nil {
					return fmt.Errorf("failed to move extracted content: %w", err)
				}
			}
			return nil
		}
	}

	// Otherwise, move all extracted content to destination
	if err := os.MkdirAll(fullDestPath, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}
	for _, entry := range entries {
		src := filepath.Join(tempPath, entry.Name())
		dst := filepath.Join(fullDestPath, entry.Name())
		if err := os.Rename(src, dst); err != nil {
			return fmt.Errorf("failed to move extracted content: %w", err)
		}
	}

	return nil
}

// Exists checks if the given path exists in storage (file or directory).
// path should be relative to BaseDir.
func (e *FileEngine) Exists(ctx context.Context, path string) bool {
	fullPath := filepath.Join(e.BaseDir, path)
	_, err := os.Stat(fullPath)
	return err == nil
}

// Remove cleans up assets from the disk at the given path.
// path should be relative to BaseDir.
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
