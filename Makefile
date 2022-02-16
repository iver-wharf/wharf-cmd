.PHONY: install tidy deps \
	proto lint lint-md \
	lint-go lint-fix lint-md-fix

ifeq ($(OS),Windows_NT)
wharf.exe:
	go build -o wharf.exe
else
wharf:
	go build -o wharf
endif

install:
	go install

tidy:
	go mod tidy

deps:
	go install github.com/mgechev/revive@latest
	go install golang.org/x/tools/cmd/goimports@latest
	go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.26
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.1
	npm install

proto:
	protoc --go_out=. --go_opt=paths=source_relative \
	--go-grpc_out=. --go-grpc_opt=paths=source_relative \
	./api/workerapi/v1/rpc/worker.proto

lint: lint-md lint-go
lint-fix: lint-md-fix

lint-md:
	npx remark . .github

lint-md-fix:
	npx remark . .github -o

lint-go:
	revive -formatter stylish -config revive.toml ./...
