package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// Version of the utility
const Version = "1.2.1"

// CopyStats tracks statistics during the copy operation
type CopyStats struct {
	FilesCopied  int
	FilesSkipped int
	BytesCopied  int64
	ExtraFound   int
	ExtraDeleted int
	ExtraBytes   int64
	StartTime    time.Time
}

// SyncOptions holds the synchronization configuration
type SyncOptions struct {
	DetectExtra bool
	DeleteExtra bool
}

// sanitizeFATTime clamps timestamps to the valid FAT/exFAT range to avoid invalid-date failures.
// FAT/exFAT valid range is approximately 1980-01-01 00:00:00 to 2107-12-31 23:59:58 (2-second resolution).
func sanitizeFATTime(t time.Time) time.Time {
	min := time.Date(1980, time.January, 1, 0, 0, 0, 0, time.Local)
	max := time.Date(2107, time.December, 31, 23, 59, 58, 0, time.Local)
	if t.Before(min) {
		return min
	}
	if t.After(max) {
		return max
	}
	return t
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	var detectExtra = flag.Bool("d", false, "detect extra files in destination not present in source")
	var deleteExtra = flag.Bool("D", false, "detect and delete extra files in destination not present in source")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] <source1> [source2...] <destination>\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nOptions:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s source dest              # Basic copy\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -d source dest           # Copy and detect extra files\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -D source dest           # Copy and delete extra files\n", os.Args[0])
	}

	flag.Parse()
	args := flag.Args()

	if len(args) < 2 {
		flag.Usage()
		fmt.Printf("Version: %s\n", Version)
		return fmt.Errorf("insufficient arguments")
	}

	syncOptions := &SyncOptions{
		DetectExtra: *detectExtra || *deleteExtra, // -D implies -d
		DeleteExtra: *deleteExtra,
	}

	// Last argument is destination, everything else is sources
	sources := args[:len(args)-1]
	destination := args[len(args)-1]

	// Validate all sources exist
	for _, source := range sources {
		if _, err := os.Stat(source); os.IsNotExist(err) {
			return fmt.Errorf("source '%s' does not exist", source)
		} else if err != nil {
			return fmt.Errorf("failed to get source info for '%s': %w", source, err)
		}
	}

	// Check if destination exists and is a directory
	destInfo, destErr := os.Stat(destination)
	isDestDir := destErr == nil && destInfo.IsDir()

	// For multiple sources, destination must be a directory (or will be created as one)
	if len(sources) > 1 && destErr == nil && !isDestDir {
		return fmt.Errorf("when copying multiple sources, destination must be a directory")
	}

	// Initialize statistics
	stats := &CopyStats{
		StartTime: time.Now(),
	}

	// Copy each source
	for _, source := range sources {
		var targetPath string

		if len(sources) == 1 {
			// Single source: use standard cp behavior
			if isDestDir {
				// Destination exists and is directory: put source inside it
				srcName := filepath.Base(source)
				targetPath = filepath.Join(destination, srcName)
			} else {
				// Destination doesn't exist or is file: use as-is
				targetPath = destination
			}
		} else {
			// Multiple sources: always put inside destination directory
			if destErr != nil {
				// Destination doesn't exist, create it as directory
				if err := os.MkdirAll(destination, 0755); err != nil {
					return fmt.Errorf("failed to create destination directory '%s': %w", destination, err)
				}
			}
			srcName := filepath.Base(source)
			targetPath = filepath.Join(destination, srcName)
		}

		if err := copyRecursively(source, targetPath, stats); err != nil {
			return err
		}
	}

	// Handle extra file detection/deletion for single source scenarios
	if len(sources) == 1 && syncOptions.DetectExtra {
		source := sources[0]
		var finalDestination string

		if isDestDir {
			// Source was copied into the destination directory
			srcName := filepath.Base(source)
			finalDestination = filepath.Join(destination, srcName)
		} else {
			// Source was copied as the destination
			finalDestination = destination
		}

		if err := handleExtraFiles(source, finalDestination, syncOptions, stats); err != nil {
			return err
		}
	}

	// Display summary statistics
	showSummary(stats, syncOptions)
	return nil
}

// copyRecursively copies files and directories from src to dst recursively
func copyRecursively(src, dst string, stats *CopyStats) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("failed to get source info: %w", err)
	}

	if srcInfo.IsDir() {
		return copyDirectory(src, dst, srcInfo, stats)
	}
	return copyFile(src, dst, srcInfo, stats)
}

// copyDirectory creates the destination directory and copies all contents
func copyDirectory(src, dst string, srcInfo os.FileInfo, stats *CopyStats) error {
	// Create destination directory with same permissions
	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return fmt.Errorf("failed to create directory '%s': %w", dst, err)
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

		if err := copyRecursively(srcPath, dstPath, stats); err != nil {
			return err
		}
	}

	// After all contents are copied, set directory times to a sanitized source time
	m := sanitizeFATTime(srcInfo.ModTime())
	if err := os.Chtimes(dst, m, m); err != nil {
		return fmt.Errorf("failed to set directory times for '%s': %w", dst, err)
	}

	return nil
}

// formatBytes formats bytes with appropriate prefixes
func formatBytes(bytes int64) string {
	if bytes >= 1e9 {
		return fmt.Sprintf("%.1fGB", float64(bytes)/1e9)
	} else if bytes >= 1e6 {
		return fmt.Sprintf("%.0fMB", float64(bytes)/1e6)
	} else if bytes >= 1e3 {
		return fmt.Sprintf("%.0fkB", float64(bytes)/1e3)
	} else {
		return fmt.Sprintf("%dB", bytes)
	}
}

// formatSpeed formats bytes per second with appropriate prefixes
func formatSpeed(bytesPerSec float64) string {
	if bytesPerSec >= 1e9 {
		return fmt.Sprintf("%.1fGB/s", bytesPerSec/1e9)
	} else if bytesPerSec >= 1e6 {
		return fmt.Sprintf("%.0fMB/s", bytesPerSec/1e6)
	} else if bytesPerSec >= 1e3 {
		return fmt.Sprintf("%.0fkB/s", bytesPerSec/1e3)
	} else {
		return fmt.Sprintf("%.0fB/s", bytesPerSec)
	}
}

// handleExtraFiles handles detection and optional deletion of extra files in destination
func handleExtraFiles(src, dst string, syncOptions *SyncOptions, stats *CopyStats) error {
	// Build a map of all files/directories that should exist in destination
	sourceItems := make(map[string]bool)

	srcInfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("failed to stat source '%s': %w", src, err)
	}

	if srcInfo.IsDir() {
		err = filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// Get relative path from source root
			relPath, err := filepath.Rel(src, path)
			if err != nil {
				return err
			}

			// Skip the root directory itself
			if relPath == "." {
				return nil
			}

			sourceItems[relPath] = true
			return nil
		})
		if err != nil {
			return fmt.Errorf("failed to walk source directory '%s': %w", src, err)
		}
	} else {
		// For single files, we just check if the destination file matches
		return nil // No extra files to handle for single file copy
	}

	// Now check destination for extra files
	var extraFiles []string
	var extraDirs []string

	err = filepath.Walk(dst, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// If we can't access a file, skip it but don't fail
			return nil
		}

		// Get relative path from destination root
		relPath, err := filepath.Rel(dst, path)
		if err != nil {
			return err
		}

		// Skip the root directory itself
		if relPath == "." {
			return nil
		}

		// Check if this item exists in source
		if !sourceItems[relPath] {
			if info.IsDir() {
				extraDirs = append(extraDirs, path)
				// Skip walking inside this directory since we'll delete it entirely
				return filepath.SkipDir
			} else {
				extraFiles = append(extraFiles, path)

				// Add to statistics
				stats.ExtraFound++
				stats.ExtraBytes += info.Size()
			}
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to walk destination directory '%s': %w", dst, err)
	}

	// Add directory statistics
	for range extraDirs {
		stats.ExtraFound++
	}

	// Report extra files found
	if len(extraFiles) > 0 || len(extraDirs) > 0 {
		fmt.Printf("\nExtra files/directories found in destination:\n")
		for _, file := range extraFiles {
			fmt.Printf("  FILE: %s\n", file)
		}
		for _, dir := range extraDirs {
			fmt.Printf("  DIR:  %s\n", dir)
		}
	}

	// Delete if requested
	if syncOptions.DeleteExtra {
		if len(extraFiles) > 0 || len(extraDirs) > 0 {
			fmt.Printf("\nDeleting extra files/directories...\n")
		}

		// Delete files first
		for _, file := range extraFiles {
			if err := os.Remove(file); err != nil {
				fmt.Printf("  WARNING: Failed to delete file '%s': %v\n", file, err)
			} else {
				fmt.Printf("  DELETED: %s\n", file)
				stats.ExtraDeleted++
			}
		}

		// Delete directories (they should be empty after deleting files)
		for _, dir := range extraDirs {
			if err := os.RemoveAll(dir); err != nil {
				fmt.Printf("  WARNING: Failed to delete directory '%s': %v\n", dir, err)
			} else {
				fmt.Printf("  DELETED: %s\n", dir)
				stats.ExtraDeleted++
			}
		}
	}

	return nil
}

// showSummary displays the final statistics
func showSummary(stats *CopyStats, syncOptions *SyncOptions) {
	totalTime := time.Since(stats.StartTime)
	overallSpeed := float64(stats.BytesCopied) / totalTime.Seconds()

	fmt.Printf("\nSummary: %d files copied, %d files skipped, %s copied in %v (%s)",
		stats.FilesCopied,
		stats.FilesSkipped,
		formatBytes(stats.BytesCopied),
		totalTime.Round(time.Millisecond),
		formatSpeed(overallSpeed))

	// Add extra files information if sync options are enabled
	if syncOptions.DetectExtra {
		if syncOptions.DeleteExtra {
			fmt.Printf(", %d extra items deleted", stats.ExtraDeleted)
		} else {
			fmt.Printf(", %d extra items found", stats.ExtraFound)
		}

		if stats.ExtraBytes > 0 {
			fmt.Printf(" (%s)", formatBytes(stats.ExtraBytes))
		}
	}

	fmt.Printf("\n")
}

// copyFile copies a single file from src to dst if needed
func copyFile(src, dst string, srcInfo os.FileInfo, stats *CopyStats) error {
	// Check if we need to copy the file
	needsCopy, err := needsUpdate(src, dst, srcInfo)
	if err != nil {
		return err
	}

	if !needsCopy {
		fmt.Printf("%s (skipped - up to date)\n", src)
		stats.FilesSkipped++
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
	// We'll close explicitly before setting timestamps to avoid Windows resetting mtime on Close

	// Copy file contents and measure time
	startTime := time.Now()
	bytesWritten, err := io.Copy(dstFile, srcFile)
	elapsedTime := time.Since(startTime)
	if err != nil {
		return fmt.Errorf("failed to copy file content from '%s' to '%s': %w", src, dst, err)
	}

	// Ensure data is flushed to disk and close the handle before setting timestamps.
	if err := dstFile.Sync(); err != nil {
		return fmt.Errorf("failed to flush destination file '%s': %w", dst, err)
	}
	if err := dstFile.Close(); err != nil {
		return fmt.Errorf("failed to close destination file '%s': %w", dst, err)
	}

	// Set file times to match source AFTER the writing handle is closed, using sanitized time.
	m := sanitizeFATTime(srcInfo.ModTime())
	if err := os.Chtimes(dst, m, m); err != nil {
		return fmt.Errorf("failed to set file times for '%s': %w", dst, err)
	}

	// Calculate and display speed
	elapsedSeconds := elapsedTime.Seconds()
	if elapsedSeconds < 0.001 { // Minimum 1ms to avoid division by near-zero
		elapsedSeconds = 0.001
	}
	speed := float64(bytesWritten) / elapsedSeconds
	fmt.Printf(" (%d bytes, %s)\n", bytesWritten, formatSpeed(speed))

	// Update statistics
	stats.FilesCopied++
	stats.BytesCopied += bytesWritten
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

	// Compare modification times with 5-second tolerance for filesystems like exFAT
	// which have 2-second resolution (we use 5 seconds for safety margin)
	srcModTime := srcInfo.ModTime()
	dstModTime := dstInfo.ModTime()

	timeDiff := srcModTime.Sub(dstModTime)
	if timeDiff < 0 {
		timeDiff = -timeDiff
	}

	// If the time difference is more than 5 seconds, consider it different
	if timeDiff > 5*time.Second {
		return true, nil
	}

	// Files are the same size and have similar modification times
	return false, nil
}
