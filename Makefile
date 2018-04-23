
build:
	docker build -t line_server .

line_server: lines.go line_server.go
	go build -o line_server .
