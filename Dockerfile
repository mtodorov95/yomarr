# Frontend
FROM node:24-alpine AS fe-builder
WORKDIR /web
COPY web/package*.json ./
RUN npm install
COPY web/ .
RUN npm run build

# Backend
FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod ./
COPY . .
COPY --from=fe-builder /web/dist ./web/dist
RUN CGO_ENABLED=0 GOOS=linux go build -o yomarr main.go

# Final
FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/yomarr .

EXPOSE 8080
CMD ["./yomarr"]
