all: todo
	@go vet
	@golint .

run:
	@go run server.go config.go datasource.go rdfstore.go handlers.go queue.go

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
	@wget http://necolas.github.com/normalize.css/3.0.1/normalize.css -O data/public/css/normalize.css
	@wget http://cdn.ractivejs.org/latest/ractive.min.js -O data/public/js/ractive.js
	@wget https://raw.github.com/ractivejs/ractive-events-keys/master/ractive-events-keys.js -O data/public/js/ractive-events-keys.js
	@wget http://underscorejs.org/underscore-min.js -O data/public/js/underscore-min.js

build: deps
	@export GOBIN=$(shell pwd)
	@go build
