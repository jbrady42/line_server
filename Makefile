
build: bin/line_server
	docker build -t line_server .

bin/line_server: lines.go line_server.go
	go build -o bin/line_server .
