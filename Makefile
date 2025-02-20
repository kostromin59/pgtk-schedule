test-integration:
	go tool goose postgres "$(CONN)" up -dir migrations
	DB_CONN="$(CONN)" go test -tags=integration ./...
	