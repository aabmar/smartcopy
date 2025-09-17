SmartCopy is a command line utility designed to copy files and directories recursively from one location to another. It will skip files with the same size and modification date to optimize the copying process. This way it can be used to update the destination with files that have changed or are new. Due to this, the file dates has to be copied as well.

It is written in Go, which makes it cross-platform and easy to compile into a single binary. It will be optimized so that it is just as fast as the native cp and copy commands.

If it experience an error, it will stop and give a good description of the error.

Please read the /README.md file for more information about current status. Please update the README.md file when you make changes that should be documented. Please remove all info from README.md that is not relevant anymore. By keeping it updated on every task we do, it will be current.

If you do not complete a task in one batch, please update the TODO list in README.md to reflect what is left to do. It is important to not forget if we leave something of the requested task undone.

The main file is in main.go

The tests are in test/main.go


Please update the test file, not add new test files or batch files.

- You can run the test file with "make test"
- you can build the main file with "make build"
