.PHONY: all
all:
	@mkdir -p bin/ || true 
	go build -o bin/v1-migration -ldflags="-w -s"

.PHONY: tests
tests: all
	bin/v1-migration tests/