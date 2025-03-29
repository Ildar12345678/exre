# FROM golang:1.22-alpine AS builder
FROM golang:alpine3.18 AS builder

# Install required dependencies
RUN apk add --no-cache gcc musl-dev sqlite-dev gcompat

# Set working directory
WORKDIR /app

# Copy go mod files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application
COPY main.go /app/main.go
COPY static/ /app/static/
COPY internal/ /app/internal/

# Enable CGO and build the application
ENV CGO_ENABLED=1
RUN go build -o app ./

# Create a lightweight runtime image
FROM alpine:latest

# Install SQLite runtime
RUN apk add --no-cache sqlite-libs

# Set working directory
WORKDIR /app

# Copy the built application from builder
COPY --from=builder /app/app .
COPY --from=builder /app/static/ /app/static/

# Set executable permission (if needed)
RUN chmod +x app

EXPOSE 5000

# Command to run the application
CMD ["./app"]
