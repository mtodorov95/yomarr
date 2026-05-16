FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod ./
# RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o yomarr ./cmd/yomarr/main.go

FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app/yomarr .

EXPOSE 8080

CMD ["./yomarr"]
