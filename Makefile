.PHONY: install tidy deps \
	lint lint-md lint-go \
	lint-fix lint-md-fix

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
	npm install

lint: lint-md lint-go
lint-fix: lint-md-fix

lint-md:
	npx remark . .github

lint-md-fix:
	npx remark . .github -o

lint-go:
	revive -formatter stylish -config revive.toml ./...