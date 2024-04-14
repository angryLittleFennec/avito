# Use the official Golang image
FROM golang:1.21-alpine3.19 AS builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies
RUN go mod download

# Copy the source code into the container
COPY . .

# Build the Go app
RUN go build -o main .

RUN chmod +x "./main"

EXPOSE 8080

# Command to run the executable
CMD ["./main"]
