run:
	set -a; . ./.env; set +a; go run ./cmd/api

migrate:
	set -a; . ./.env; set +a; go run ./cmd/api migrate

test:
	go test ./...

check:
	go build ./... && go vet ./... && go test ./...
