migrate:
	go run ./cmd/migrator/main.go --migrations-path=./migrations --db-user=fin_admin --db-password=12345678FinAdmin --db-host=localhost --db-port=5432 --db-name=finaptica

sso:
	go run ./cmd/sso/main.go --config=./config/local.yaml