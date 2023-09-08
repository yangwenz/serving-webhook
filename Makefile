
server:
	go run main.go

test:
	go test -v -cover -short ./...

mock:
	mockgen -package mockstore -destination storage/mock/store.go github.com/yangwenz/model-webhook/storage Store
	mockgen -package mockstore -destination storage/mock/cache.go github.com/yangwenz/model-webhook/storage Cache

docker:
	docker build -t yangwenz/model-webhook:v1 .
	docker push yangwenz/model-webhook:v1

.PHONY: server test mock