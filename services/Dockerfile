# build
FROM golang:1.24.5 AS builder

WORKDIR /app

# caching dependencies
COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o app cmd/main.go

# final
FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/app .
COPY --from=builder /app/config/config.yml .

CMD [ "./app" ]
