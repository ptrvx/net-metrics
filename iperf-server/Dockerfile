# Build stage
FROM golang:1.22 AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

ENV GOFLAGS=-mod=vendor

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o iperf-server .

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/iperf-server .

EXPOSE 5201

EXPOSE 8080

CMD ["./iperf-server"]
