FROM golang:alpine3.18 AS builder

RUN apk add --no-cache gcc musl-dev sqlite-dev gcompat

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY main.go /app/main.go
COPY static/ /app/static/
COPY internal/ /app/internal/

ENV CGO_ENABLED=1
RUN go build -o app ./

FROM alpine:latest

RUN apk add --no-cache sqlite-libs

WORKDIR /app

COPY --from=builder /app/app .
COPY --from=builder /app/static/ /app/static/

RUN chmod +x app

EXPOSE 5000

CMD ["./app"]
