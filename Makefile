.PHONY: test

check-updates:
	go test -tags check_updates ./resource

test:
	go test ./...
