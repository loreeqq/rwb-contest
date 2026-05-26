.PHONY: test test-race cover gcov_report clean

test:
	go test ./internal/storage ./internal/service

test-race:
	go test -race ./internal/storage ./internal/service

gcov_report:
	go test \
		-coverpkg=./internal/storage,./internal/service \
		-coverprofile=coverage.out \
		./internal/storage ./internal/service

	go tool cover -html=coverage.out -o coverage.html
	open coverage.html

clean:
	rm -f coverage.out coverage.html
