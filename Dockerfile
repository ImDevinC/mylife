FROM golang:1.19-alpine

WORKDIR /usr/src/app

COPY . .

RUN go build -o app ./cmd/bot/main.go

ENTRYPOINT ["/usr/src/app/app"]