FROM golang:1.23.3-alpine3.20 as builder

# Install necessary packages
RUN apk update && apk add --no-cache \
    git \
    openssh \
    tzdata \
    build-base \
    python3 \
    net-tools

WORKDIR /app

# Copy go.mod and go.sum first
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the rest of the files
COPY . .
COPY .env.example .env

# Install gin
RUN go install github.com/buu700/gin@latest

# Now run go mod tidy
RUN go mod tidy

# Build the application
RUN make build

# Final stage
FROM alpine:latest
RUN apk update && apk upgrade && \
    apk --update --no-cache add tzdata && \
    apk --no-cache add curl && \
    mkdir /app
WORKDIR /app
EXPOSE 8002
COPY --from=builder /app /app
ENTRYPOINT ["/app/field-service"]