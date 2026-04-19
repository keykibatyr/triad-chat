FROM golang:tip-alpine3.23 AS builder

#create a working directory/folder inside the container
WORKDIR /app

# Copy Go modules and dependencies
COPY go.mod go.sum ./
RUN go mod download
RUN go install github.com/air-verse/air@latest

EXPOSE 8080
