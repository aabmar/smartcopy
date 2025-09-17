package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	args := os.Args[1:]
	if len(args) != 2 {
		return fmt.Errorf("usage: smartcopy <from> <to>")
	}

	from := args[0]
	to := args[1]

	// Validate source exists
	if _, err := os.Stat(from); os.IsNotExist(err) {
		return fmt.Errorf("source '%s' does not exist", from)
	}

	return copyRecursively(from, to)
}

// copyRecursively copies files and directories from src to dst recursively
func copyRecursively(src, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("failed to get source info: %w", err)
	}

	if srcInfo.IsDir() {
		return copyDirectory(src, dst, srcInfo)
	}
	return copyFile(src, dst, srcInfo)
}

// copyDirectory creates the destination directory and copies all contents
func copyDirectory(src, dst string, srcInfo os.FileInfo) error {
	// Create destination directory with same permissions
	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return fmt.Errorf("failed to create directory '%s': %w", dst, err)
	}

	// Copy directory modification time
	if err := os.Chtimes(dst, srcInfo.ModTime(), srcInfo.ModTime()); err != nil {
		return fmt.Errorf("failed to set directory times for '%s': %w", dst, err)
	}

	// Read directory entries
	entries, err := os.ReadDir(src)
	if err != nil {
		return fmt.Errorf("failed to read directory '%s': %w", src, err)
	}

	// Copy each entry recursively
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if err := copyRecursively(srcPath, dstPath); err != nil {
			return err
		}
	}

	return nil
}

// copyFile copies a single file from src to dst if needed
func copyFile(src, dst string, srcInfo os.FileInfo) error {
	// Check if we need to copy the file
	needsCopy, err := needsUpdate(src, dst, srcInfo)
	if err != nil {
		return err
	}

	if !needsCopy {
		fmt.Printf("%s (skipped - up to date)\n", src)
		return nil
	}

	fmt.Printf("%s", src)

	// Create destination directory if it doesn't exist
	dstDir := filepath.Dir(dst)
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory '%s': %w", dstDir, err)
	}

	// Open source file
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file '%s': %w", src, err)
	}
	defer srcFile.Close()

	// Create destination file
	dstFile, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, srcInfo.Mode())
	if err != nil {
		return fmt.Errorf("failed to create destination file '%s': %w", dst, err)
	}
	defer dstFile.Close()

	// Copy file contents
	bytesWritten, err := io.Copy(dstFile, srcFile)
	if err != nil {
		return fmt.Errorf("failed to copy file content from '%s' to '%s': %w", src, dst, err)
	}

	// Set file modification time to match source
	if err := os.Chtimes(dst, srcInfo.ModTime(), srcInfo.ModTime()); err != nil {
		return fmt.Errorf("failed to set file times for '%s': %w", dst, err)
	}

	fmt.Printf(" (%d bytes)\n", bytesWritten)
	return nil
}

// needsUpdate checks if the destination file needs to be updated
func needsUpdate(src, dst string, srcInfo os.FileInfo) (bool, error) {
	dstInfo, err := os.Stat(dst)
	if os.IsNotExist(err) {
		// Destination doesn't exist, needs copy
		return true, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to get destination file info for '%s': %w", dst, err)
	}

	// Compare size and modification time
	if srcInfo.Size() != dstInfo.Size() {
		return true, nil
	}

	// Compare modification times (truncate to second precision for cross-platform compatibility)
	srcModTime := srcInfo.ModTime().Truncate(time.Second)
	dstModTime := dstInfo.ModTime().Truncate(time.Second)

	if !srcModTime.Equal(dstModTime) {
		return true, nil
	}

	// Files are the same size and have the same modification time
	return false, nil
}
