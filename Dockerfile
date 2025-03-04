FROM golang:1.22-alpine AS builder
WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .
RUN go build -o server .

FROM alpine:latest  
WORKDIR /root/

COPY --from=builder /app/server/doctor-aibolit .

EXPOSE 8080
CMD ["./doctor-aibolit"]
