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

	// Test 8: Copy directory into existing directory (new cp-like behavior)
	fmt.Println("\n11. Test 8: Copy directory into existing directory")
	if err := os.MkdirAll(joinRoot("existing_dir"), 0755); err != nil {
		return fmt.Errorf("failed to create existing directory: %w", err)
	}
	fmt.Println("Running: smartcopy test_src existing_dir (should create existing_dir/test_src/)")
	if err := runSmartcopy(joinRoot("smartcopy.exe"), joinRoot("test_src"), joinRoot("existing_dir")); err != nil {
		return fmt.Errorf("copy to existing directory failed: %w", err)
	}

	// Verify the directory structure is correct
	if err := verifyDirectoryStructure(joinRoot("existing_dir", "test_src"), joinRoot); err != nil {
		return fmt.Errorf("directory structure verification failed: %w", err)
	}

	// Test 9: Copy file into existing directory
	fmt.Println("\n12. Test 9: Copy file into existing directory")
	if err := os.MkdirAll(joinRoot("file_dest_dir"), 0755); err != nil {
		return fmt.Errorf("failed to create file destination directory: %w", err)
	}
	fmt.Println("Running: smartcopy test_src/file1.txt file_dest_dir (should create file_dest_dir/file1.txt)")
	if err := runSmartcopy(joinRoot("smartcopy.exe"), joinRoot("test_src", "file1.txt"), joinRoot("file_dest_dir")); err != nil {
		return fmt.Errorf("copy file to existing directory failed: %w", err)
	}

	// Verify the file was placed correctly
	expectedFile := joinRoot("file_dest_dir", "file1.txt")
	if _, err := os.Stat(expectedFile); os.IsNotExist(err) {
		return fmt.Errorf("expected file %s was not created", expectedFile)
	}
	fmt.Printf("  ✓ Verified: File correctly placed at %s\n", expectedFile)

	// Test 10: Copy to non-existing path (should create with source name)
	fmt.Println("\n13. Test 10: Copy to non-existing path")
	fmt.Println("Running: smartcopy test_src/file1.txt new_file_copy.txt")
	if err := runSmartcopy(joinRoot("smartcopy.exe"), joinRoot("test_src", "file1.txt"), joinRoot("new_file_copy.txt")); err != nil {
		return fmt.Errorf("copy to new path failed: %w", err)
	}

	// Verify the file was created with the specified name
	expectedNewFile := joinRoot("new_file_copy.txt")
	if _, err := os.Stat(expectedNewFile); os.IsNotExist(err) {
		return fmt.Errorf("expected file %s was not created", expectedNewFile)
	}
	fmt.Printf("  ✓ Verified: File correctly created as %s\n", expectedNewFile)

	// Test 11: Multiple sources to existing directory
	fmt.Println("\n14. Test 11: Multiple sources to existing directory")
	if err := createMultipleTestSources(joinRoot); err != nil {
		return fmt.Errorf("failed to create multiple test sources: %w", err)
	}
	if err := os.MkdirAll(joinRoot("multi_dest"), 0755); err != nil {
		return fmt.Errorf("failed to create multi destination directory: %w", err)
	}
	fmt.Println("Running: smartcopy src1 src2 src3 multi_dest")
	if err := runMultiSmartcopy(joinRoot("smartcopy.exe"), []string{
		joinRoot("src1"), joinRoot("src2"), joinRoot("src3"), joinRoot("multi_dest"),
	}); err != nil {
		return fmt.Errorf("multiple sources copy failed: %w", err)
	}

	// Verify all sources were copied into destination
	expectedPaths := []string{
		joinRoot("multi_dest", "src1"),
		joinRoot("multi_dest", "src2"),
		joinRoot("multi_dest", "src3"),
	}
	for _, path := range expectedPaths {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return fmt.Errorf("expected path %s was not created", path)
		}
	}
	fmt.Printf("  ✓ Verified: All sources correctly copied to multi_dest\n")

	// Test 12: Multiple sources to non-existing directory (should create it)
	fmt.Println("\n15. Test 12: Multiple sources to non-existing directory")
	fmt.Println("Running: smartcopy src1 src2 new_multi_dest")
	if err := runMultiSmartcopy(joinRoot("smartcopy.exe"), []string{
		joinRoot("src1"), joinRoot("src2"), joinRoot("new_multi_dest"),
	}); err != nil {
		return fmt.Errorf("multiple sources to new directory failed: %w", err)
	}

	// Verify destination was created and sources copied
	newExpectedPaths := []string{
		joinRoot("new_multi_dest", "src1"),
		joinRoot("new_multi_dest", "src2"),
	}
	for _, path := range newExpectedPaths {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return fmt.Errorf("expected path %s was not created", path)
		}
	}
	fmt.Printf("  ✓ Verified: Multiple sources copied to new directory\n")

	// Test 13: Multiple sources to existing file (should fail)
	fmt.Println("\n16. Test 13: Multiple sources to existing file (should fail)")
	if err := createFile(joinRoot("existing_file.txt"), "existing file"); err != nil {
		return fmt.Errorf("failed to create existing file: %w", err)
	}
	fmt.Println("Running: smartcopy src1 src2 existing_file.txt (should fail)")
	cmd2 := exec.Command(joinRoot("smartcopy.exe"), joinRoot("src1"), joinRoot("src2"), joinRoot("existing_file.txt"))
	output2, err2 := cmd2.CombinedOutput()
	if err2 == nil {
		return fmt.Errorf("expected error when copying multiple sources to file, but command succeeded")
	}
	fmt.Printf("  Expected error output: %s", string(output2))

	// Clean up test directories
	fmt.Println("\n17. Cleaning up test directories...")
	cleanupTestDirs(joinRoot)
	os.RemoveAll(joinRoot("existing_dir"))
	os.RemoveAll(joinRoot("file_dest_dir"))
	os.RemoveAll(joinRoot("new_file_copy.txt"))
	os.RemoveAll(joinRoot("src1"))
	os.RemoveAll(joinRoot("src2"))
	os.RemoveAll(joinRoot("src3"))
	os.RemoveAll(joinRoot("multi_dest"))
	os.RemoveAll(joinRoot("new_multi_dest"))
	os.RemoveAll(joinRoot("existing_file.txt"))

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

func verifyDirectoryStructure(basePath string, joinRoot func(parts ...string) string) error {
	// Check that the expected files exist in the copied directory structure
	expectedFiles := []string{
		filepath.Join(basePath, "file1.txt"),
		filepath.Join(basePath, "file2.txt"),
		filepath.Join(basePath, "small.txt"),
		filepath.Join(basePath, "large.dat"),
		filepath.Join(basePath, "subdir", "nested.txt"),
		filepath.Join(basePath, "subdir", "nested2.txt"),
	}

	for _, file := range expectedFiles {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			return fmt.Errorf("expected file %s does not exist", file)
		}
	}

	fmt.Printf("  ✓ Verified: Directory structure correctly created at %s\n", basePath)
	return nil
}

func createMultipleTestSources(joinRoot func(parts ...string) string) error {
	// Create src1 directory with files
	if err := os.MkdirAll(joinRoot("src1"), 0755); err != nil {
		return err
	}
	if err := createFile(joinRoot("src1", "file_a.txt"), "Content of file A in src1"); err != nil {
		return err
	}
	if err := createFile(joinRoot("src1", "file_b.txt"), "Content of file B in src1"); err != nil {
		return err
	}

	// Create src2 directory with files
	if err := os.MkdirAll(joinRoot("src2"), 0755); err != nil {
		return err
	}
	if err := createFile(joinRoot("src2", "doc.txt"), "Document in src2"); err != nil {
		return err
	}

	// Create src3 as a single file
	if err := createFile(joinRoot("src3"), "This is src3 as a file"); err != nil {
		return err
	}

	return nil
}

func runMultiSmartcopy(binPath string, args []string) error {
	cmd := exec.Command(binPath, args...)
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
