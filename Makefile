all: todo
	@go vet
	@golint .

run:
	@go run server.go config.go datasource.go rdfstore.go handlers.go queue.go indexing.go ids.go rdf2marc.go sync.go queries.go

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
	@wget http://necolas.github.com/normalize.css/3.0.1/normalize.css -O data/public/css/normalize.css
	@wget http://cdn.ractivejs.org/edge/ractive.min.js -O data/public/js/ractive.js
	@wget https://github.com/ractivejs/cdn.ractivejs.org/blob/gh-pages/0.7.3/ractive.min.js.map -O data/public/js/ractive.js.map
	@wget https://raw.github.com/ractivejs/ractive-events-keys/master/dist/ractive-events-keys.js -O data/public/js/ractive-events-keys.js
	@wget http://underscorejs.org/underscore-min.js -O data/public/js/underscore-min.js

build:
	@export GOBIN=$(shell pwd)
	@go build

test:
	@go get ./...
	@go list -f '{{range .TestImports}}{{.}} {{end}}' ./... | xargs -n1 go get -d
	@go test

indexes:
	@curl -XDELETE http://localhost:9200/public
	@curl -XPUT http://localhost:9200/public -d @data/es_settings.json

mappings:
	@go run setupmappings.go indexing.go

docker-up:
	@vagrant up && vagrant provision

docker-stop:
	@vagrant ssh -c '(docker stop armillaria && docker rm armillaria) | true'

docker-restart: docker-stop
	@echo "======= RESTARTING ARMILLARIA CONTAINER======\n"
	@vagrant ssh -c 'docker run -d -e SERVER_PORT=8080 \
	-p 8080:8080 \
	--name armillaria \
	--link virtuoso:virtuoso \
	--link elasticsearch:elasticsearch \
	-t armillaria'
