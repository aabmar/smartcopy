// Test program for SmartCopy
// Run with: go run test/main.go
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

func main() {
	fmt.Println("=== SmartCopy Test Suite ===")

	if err := runTests(); err != nil {
		fmt.Fprintf(os.Stderr, "Test failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\n=== All tests completed successfully! ===")
}

func runTests() error {
	// Determine repository root based on this file's location
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		return fmt.Errorf("unable to determine runtime caller info")
	}
	repoRoot := filepath.Dir(filepath.Dir(thisFile)) // .../smartcopy/test/main.go -> .../smartcopy

	joinRoot := func(parts ...string) string {
		return filepath.Join(append([]string{repoRoot}, parts...)...)
	}

	// Clean up any existing test directories
	fmt.Println("1. Cleaning up existing test directories...")
	cleanupTestDirs(joinRoot)

	// Create test directory structure
	fmt.Println("2. Creating test source directory structure...")
	if err := createTestStructure(joinRoot); err != nil {
		return fmt.Errorf("failed to create test structure: %w", err)
	}

	// Build smartcopy if needed
	fmt.Println("3. Building smartcopy...")
	if err := buildSmartcopy(repoRoot); err != nil {
		return fmt.Errorf("failed to build smartcopy: %w", err)
	}

	// Test 1: Initial copy (all files should be copied)
	fmt.Println("\n4. Test 1: Initial copy (all files should be copied)")
	fmt.Println("Running: smartcopy test_src test_dst")
	if err := runSmartcopy(joinRoot("smartcopy.exe"), joinRoot("test_src"), joinRoot("test_dst")); err != nil {
		return fmt.Errorf("initial copy failed: %w", err)
	}

	// Test 2: Second copy (all files should be skipped)
	fmt.Println("\n5. Test 2: Second copy (all files should be skipped)")
	fmt.Println("Running: smartcopy test_src test_dst")
	if err := runSmartcopy(joinRoot("smartcopy.exe"), joinRoot("test_src"), joinRoot("test_dst")); err != nil {
		return fmt.Errorf("second copy failed: %w", err)
	}

	// Test 3: Modify a file and copy again
	fmt.Println("\n6. Test 3: Modifying file1.txt and copying again")
	if err := modifyFile(joinRoot("test_src", "file1.txt"), "This is the MODIFIED content of file1"); err != nil {
		return fmt.Errorf("failed to modify file: %w", err)
	}
	fmt.Println("Running: smartcopy test_src test_dst")
	if err := runSmartcopy(joinRoot("smartcopy.exe"), joinRoot("test_src"), joinRoot("test_dst")); err != nil {
		return fmt.Errorf("modified copy failed: %w", err)
	}

	// Test 4: Add a new file and copy again
	fmt.Println("\n7. Test 4: Adding new_file.txt and copying again")
	if err := createFile(joinRoot("test_src", "new_file.txt"), "This is a newly added file"); err != nil {
		return fmt.Errorf("failed to create new file: %w", err)
	}
	fmt.Println("Running: smartcopy test_src test_dst")
	if err := runSmartcopy(joinRoot("smartcopy.exe"), joinRoot("test_src"), joinRoot("test_dst")); err != nil {
		return fmt.Errorf("new file copy failed: %w", err)
	}

	// Test 5: Add file to subdirectory and copy again
	fmt.Println("\n8. Test 5: Adding file to subdirectory and copying again")
	if err := createFile(joinRoot("test_src", "subdir", "nested_new.txt"), "New file in subdirectory"); err != nil {
		return fmt.Errorf("failed to create nested new file: %w", err)
	}
	fmt.Println("Running: smartcopy test_src test_dst")
	if err := runSmartcopy(joinRoot("smartcopy.exe"), joinRoot("test_src"), joinRoot("test_dst")); err != nil {
		return fmt.Errorf("nested new file copy failed: %w", err)
	}

	// Test 6: Single file copy
	fmt.Println("\n9. Test 6: Single file copy")
	fmt.Println("Running: smartcopy test_src/file1.txt test_dst/single_copy.txt")
	if err := runSmartcopy(joinRoot("smartcopy.exe"), joinRoot("test_src", "file1.txt"), joinRoot("test_dst", "single_copy.txt")); err != nil {
		return fmt.Errorf("single file copy failed: %w", err)
	}

	// Test 7: Error handling - non-existent source
	fmt.Println("\n10. Test 7: Error handling (non-existent source)")
	fmt.Println("Running: smartcopy nonexistent test_dst/error (should fail)")
	cmd := exec.Command(joinRoot("smartcopy.exe"), "nonexistent", joinRoot("test_dst", "error"))
	output, err := cmd.CombinedOutput()
	if err == nil {
		return fmt.Errorf("expected error for non-existent source, but command succeeded")
	}
	fmt.Printf("  Expected error output: %s", string(output))

	return nil
}

func cleanupTestDirs(joinRoot func(parts ...string) string) {
	dirs := []string{"test_src", "test_dst"}
	for _, dir := range dirs {
		os.RemoveAll(joinRoot(dir))
	}
}

func createTestStructure(joinRoot func(parts ...string) string) error {
	// Create main test directory
	if err := os.MkdirAll(joinRoot("test_src"), 0755); err != nil {
		return err
	}

	// Create subdirectory
	if err := os.MkdirAll(joinRoot("test_src", "subdir"), 0755); err != nil {
		return err
	}

	// Create test files with different sizes
	files := map[string]string{
		joinRoot("test_src", "file1.txt"):             "This is the content of file1",
		joinRoot("test_src", "file2.txt"):             "This is the content of file2 - a bit longer than file1",
		joinRoot("test_src", "small.txt"):             "Small",
		joinRoot("test_src", "subdir", "nested.txt"):  "This file is nested in a subdirectory",
		joinRoot("test_src", "subdir", "nested2.txt"): "Another nested file with different content",
	}

	for path, content := range files {
		if err := createFile(path, content); err != nil {
			return err
		}
	}

	// Create a larger binary file for speed testing
	return createLargeFile(joinRoot("test_src", "large.dat"), 1024*1024) // 1MB
}

func createFile(path, content string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), 0644)
}

func createLargeFile(path string, size int) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write size bytes of data
	data := make([]byte, 1024) // 1KB buffer
	for i := range data {
		data[i] = byte(i % 256)
	}

	written := 0
	for written < size {
		writeSize := len(data)
		if written+writeSize > size {
			writeSize = size - written
		}
		n, err := file.Write(data[:writeSize])
		if err != nil {
			return err
		}
		written += n
	}

	return nil
}

func modifyFile(path, newContent string) error {
	// Add a small delay to ensure different modification time
	time.Sleep(10 * time.Millisecond)
	return os.WriteFile(path, []byte(newContent), 0644)
}

func buildSmartcopy(repoRoot string) error {
	cmd := exec.Command("go", "build", "-o", "smartcopy.exe")
	cmd.Dir = repoRoot // Run from repository root where go.mod is
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("build failed: %v\nOutput: %s", err, string(output))
	}
	return nil
}

func runSmartcopy(binPath, src, dst string) error {
	cmd := exec.Command(binPath, src, dst)
	output, err := cmd.CombinedOutput()

	// Always show the output (including summary)
	outputStr := string(output)
	if outputStr != "" {
		// Add indentation for better readability
		lines := strings.Split(strings.TrimSpace(outputStr), "\n")
		for _, line := range lines {
			fmt.Printf("  %s\n", line)
		}
	}

	if err != nil {
		return fmt.Errorf("smartcopy failed: %v", err)
	}

	return nil
}
