ifeq ($(OS), Windows_NT)
	DEL := del
else
	DEL := rm
endif

coverage:
	go test ./... -cover -coverprofile="coverage.out"
	go tool cover -html="coverage.out"

clean:
	$(DEL) coverage.out
	go clean
