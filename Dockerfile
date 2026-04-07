FROM golang:1.25.3 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /go_final_project .

FROM ubuntu:latest

RUN apt-get update && apt-get install -y \
    ca-certificates \
    tzdata \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY --from=builder /go_final_project /app/go_final_project

COPY web /app/web
COPY .env /app/.env

ENV TODO_PORT=7540 \
    TODO_DBFILE=scheduler.db

EXPOSE 7540

CMD ["/app/go_final_project"]