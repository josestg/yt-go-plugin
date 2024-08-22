bin/demo: $(shell find . -name '*.go') # ensure bin/demo is rebuilt when any go file changes.
	go build -o bin/demo main.go

