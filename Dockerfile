FROM golang:1.21-alpine3.19 AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN go build -o main .
RUN chmod +x "./main"

EXPOSE 8080

CMD ["./main"]
