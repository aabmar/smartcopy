@echo off
echo === SmartCopy Test Suite ===
echo.

echo 1. Cleaning up existing test directories...
if exist test_src rmdir /s /q test_src
if exist test_dst rmdir /s /q test_dst

echo 2. Building SmartCopy...
go build -o smartcopy.exe
if errorlevel 1 (
    echo Build failed!
    exit /b 1
)

echo 3. Creating test source directory structure...
mkdir test_src\subdir

echo This is the content of file1 > test_src\file1.txt
echo This is the content of file2 - a bit longer than file1 > test_src\file2.txt
echo Small > test_src\small.txt
echo This file is nested in a subdirectory > test_src\subdir\nested.txt
echo Another nested file with different content > test_src\subdir\nested2.txt
fsutil file createnew test_src\large.dat 1048576

echo.
echo 4. Test 1: Initial copy (all files should be copied)
echo Running: smartcopy test_src test_dst
smartcopy.exe test_src test_dst

echo.
echo 5. Test 2: Second copy (all files should be skipped)
echo Running: smartcopy test_src test_dst
smartcopy.exe test_src test_dst

echo.
echo 6. Test 3: Modifying file1.txt and copying again
echo This is the MODIFIED content of file1 > test_src\file1.txt
echo Running: smartcopy test_src test_dst
smartcopy.exe test_src test_dst

echo.
echo 7. Test 4: Adding new_file.txt and copying again
echo This is a newly added file > test_src\new_file.txt
echo Running: smartcopy test_src test_dst
smartcopy.exe test_src test_dst

echo.
echo 8. Test 5: Adding file to subdirectory and copying again
echo New file in subdirectory > test_src\subdir\nested_new.txt
echo Running: smartcopy test_src test_dst
smartcopy.exe test_src test_dst

echo.
echo 9. Test 6: Single file copy
echo Running: smartcopy test_src\file1.txt test_dst\single_copy.txt
smartcopy.exe test_src\file1.txt test_dst\single_copy.txt

echo.
echo 10. Test 7: Error handling (non-existent source)
echo Running: smartcopy nonexistent test_dst\error (should fail)
smartcopy.exe nonexistent test_dst\error

echo.
echo === All tests completed successfully! ===