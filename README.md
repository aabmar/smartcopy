# SmartCopy

A command line utility for efficiently copying files and directories recursively. SmartCopy optimizes the copying process by skipping files that are already up to date (same size and modification date).

## Features

- **Recursive copying**: Copies directories and all their contents
- **Smart skipping**: Only copies files that have changed (different size or modification date)
- **Filesystem compatibility**: 5-second timestamp tolerance for filesystems with limited precision (e.g., exFAT)
- **Synchronization options**: Detect and optionally delete extra files in destination
- **Cross-platform**: Written in Go for Windows, macOS, and Linux
- **Progress reporting**: Shows each file being processed and bytes transferred
- **Date preservation**: Maintains original file modification times
- **Hidden files**: Copies all files including hidden and system files
- **Error handling**: Provides clear error messages when problems occur

## Usage

```bash
# Basic copying
smartcopy [options] <source1> [source2...] <destination>

# Options:
#   -d    detect extra files in destination not present in source
#   -D    detect and delete extra files in destination not present in source
```

### Examples

```bash
# Basic copying - Copy a directory recursively
smartcopy ./source ./destination

# Copy a single file
smartcopy ./file.txt ./backup/file.txt

# Copy multiple sources to destination (great for backups)
smartcopy ./documents ./photos ./projects ./backup/

# Update an existing backup (only copies changed files)
smartcopy ./project ./backup/project

# Synchronization - Detect extra files in destination
smartcopy -d ./source ./backup

# Complete synchronization - Remove extra files from destination
smartcopy -D ./source ./backup

# Perfect for maintaining a mirror backup
smartcopy -D ./important_docs ./backup/important_docs
```

### Synchronization Features

The `-d` and `-D` flags provide powerful synchronization capabilities perfect for backup scenarios:

- **`-d` (detect)**: Lists extra files/directories in destination that don't exist in source
- **`-D` (delete)**: Same as `-d` but also deletes the extra files/directories

This ensures your backup destination stays in perfect sync with the source, removing outdated files that are no longer needed.

## Filesystem Compatibility

SmartCopy is designed to work reliably across different filesystems, including those with limited timestamp precision:

### exFAT Compatibility

Some filesystems like exFAT have limited timestamp precision (2-second resolution). To handle this, SmartCopy uses a **5-second tolerance** when comparing file modification times. This means:

- Files with modification times that differ by 5 seconds or less are considered "the same"
- Only files with modification time differences greater than 5 seconds are copied
- This prevents unnecessary copying on exFAT drives while maintaining accuracy

### Why 5 Seconds?

- exFAT has 2-second timestamp resolution
- 5-second tolerance provides a safe margin for clock drift and filesystem variations
- Ensures reliable operation across different operating systems and storage devices

This feature works transparently - no special flags or configuration needed!

## Building and Testing

Use the included Makefile for common tasks:

```bash
# Build the executable
make build

# Run tests
make test

# Clean build artifacts
make clean

# Run directly with go
make run
```

Or use Go commands directly:
```bash
go build -o smartcopy.exe
```

## Architecture

The SmartCopy utility is implemented as a single Go file (`main.go`) with well-structured functions:

### Main Components

- **`main()`** and **`run()`**: Entry point and argument validation
- **`copyRecursively()`**: Main dispatcher that determines whether to copy a file or directory
- **`copyDirectory()`**: Handles recursive directory copying with permission preservation
- **`copyFile()`**: Copies individual files with progress reporting
- **`needsUpdate()`**: Determines if a file needs copying by comparing size and modification time

### Key Features

1. **Smart Comparison**: Files are compared by size and modification time with 5-second tolerance for filesystem compatibility
2. **Filesystem Compatibility**: Works seamlessly with exFAT and other filesystems that have limited timestamp precision (2-second resolution)
3. **Progress Output**: Shows filename when starting and bytes transferred when complete
4. **Error Handling**: Comprehensive error handling with descriptive messages at each step
5. **Permission Preservation**: Maintains file and directory permissions
6. **Date Preservation**: Preserves modification times for both files and directories

### File Structure

```
├── main.go          # Complete implementation
├── go.mod          # Go module definition
├── smartcopy.exe   # Compiled binary (Windows)
└── README.md       # This documentation
```

The code is designed to be robust and handle edge cases while providing clear feedback to the user about what operations are being performed.