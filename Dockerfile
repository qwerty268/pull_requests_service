# build stage
FROM golang:1.23 as builder
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o app ./cmd/main.go

# run stage
FROM debian:stable-slim
WORKDIR /app
COPY db/migration.sql ./db/migration.sql 
COPY --from=builder /app/app .
CMD ["./app"]
