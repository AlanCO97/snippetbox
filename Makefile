build:
	@go build -o ./bin/web  ./cmd/web/ 
	@cp -r ./tls ./bin/  
run: build
	@./bin/web

PHONY: build run