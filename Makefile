DB_URL=postgresql://root:secret@localhost:5432/serving_webhook?sslmode=disable

server:
	go run main.go

test:
	go test -v -cover -short ./...

mock:
	mockgen -package mockstore -destination storage/mock/store.go github.com/HyperGAI/serving-webhook/storage Store
	mockgen -package mockstore -destination storage/mock/cache.go github.com/HyperGAI/serving-webhook/storage Cache
	mockgen -package mockdb -destination db/mock/store.go github.com/HyperGAI/serving-webhook/db/sqlc Store

docker:
	docker build --platform=linux/amd64 -t yangwenz/serving-webhook:v2 .
	docker push yangwenz/serving-webhook:v2

sqlc:
	sqlc generate

postgres:
	/etc/init.d/postgresql stop
	docker run --name postgres -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -d postgres:15.4-alpine
	# kubectl run -it --rm --image=postgres:15.4-alpine --restart=Never postgres-client -- psql postgresql://root:secret@10.33.97.5:5432/serving_webhook?sslmode=disable

droppostgres:
	docker stop postgres
	docker container rm postgres

createdb:
	docker exec -it postgres createdb --username=root --owner=root serving_webhook

dropdb:
	docker exec -it postgres dropdb serving_webhook

new_migration:
	migrate create -ext sql -dir db/migration -seq $(name)

migrateup:
	migrate -path db/migration -database "$(DB_URL)" -verbose up

migratedown:
	migrate -path db/migration -database "$(DB_URL)" -verbose down

redis:
	service redis stop
	docker run --name redis -p 6379:6379 -d redis:7.2-alpine

dropredis:
	docker stop redis
	docker container rm redis

.PHONY: server test mock docker new_migration sqlc postgres createdb dropdb migrateup migratedown droppostgres redis dropredis
