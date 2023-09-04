
server:
	go run main.go

test:
	go test -v -cover -short ./...

mock:
	mockgen -package mockstore -destination storage/mock/store.go github.com/yangwenz/model-webhook/storage Store

.PHONY: server test mock