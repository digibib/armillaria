all: todo
	@go vet
	@golint .

run:
	@go run server.go config.go datasource.go

todo:
	@grep -rn TODO *.go || true
	@grep -rn println *.go || true

clean:
	@go clean
	@rm -f *.out
	@rm -f armillaria
	@rm -f *.log
