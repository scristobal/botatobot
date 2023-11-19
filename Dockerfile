FROM golang:1.21-alpine3.18
COPY . /app
WORKDIR /app
RUN go build -o botatobot cmd/botatobot/main.go
CMD [ "./botatobot" ]