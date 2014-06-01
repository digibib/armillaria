all: todo
	@go vet
	@golint .

run:
	@go run server.go config.go datasource.go handlers.go

todo:
	@grep -rn TODO *.go || true
	@grep -rn println *.go || true

clean:
	@go clean
	@rm -f *.out
	@rm -f armillaria
	@rm -f *.log

deps:
	@go get -d -v ./...
	@go list -f '{{range .TestImports}}{{.}} {{end}}' ./... | xargs -n1 go get -d
	@wget http://cdn.ractivejs.org/latest/ractive.min.js -O data/public/js/ractive.js
	@wget https://raw.github.com/ractivejs/ractive-events-keys/master/ractive-events-keys.js -O data/public/js/ractive-events-keys.js
	@wget http://underscorejs.org/underscore-min.js -O data/public/js/underscore-min.js

build: deps
	@export GOBIN=$(shell pwd)
	@go build
