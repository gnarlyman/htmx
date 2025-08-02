FROM golang:1.24-alpine

WORKDIR /app

# Install dependencies for building and watching
RUN apk add --no-cache git gcc musl-dev

# Install reflex for Go hot reloading and templ CLI for template generation/proxy
RUN go install github.com/cespare/reflex@latest
RUN go install github.com/a-h/templ/cmd/templ@latest

# Copy go mod files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# The rest of the code will be mounted via volumes in docker-compose