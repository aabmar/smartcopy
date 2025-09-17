

build: smartcopy.exe
	@echo "Build complete."

smartcopy.exe: main.go
	go build -o smartcopy.exe main.go

clean:
	rm -f smartcopy.exe

run: 
	go run main.go

test:
	go run test/main.go

.PHONY: clean run test
