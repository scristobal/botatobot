build:
	go build -o build/botatobot cmd/botatobot/main.go
run: build
	./build/botatobot