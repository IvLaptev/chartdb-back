gen: generate

generate:
	buf dep update
	buf generate

migrate:
	pipx run yandex-pgmigrate migrate -c 'host=localhost port=5432 dbname=chartdb user=chartdb password=secretpass' -d . -t latest -v -a afterAll:./migrations/grants
