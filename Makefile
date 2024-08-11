build:
	GOOS=windows CGO_ENABLED=0 go build -o uboa.exe main.go

build-linux:
	GOOS=linux CGO_ENABLED=0 go build -o uboa main.go
