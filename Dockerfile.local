FROM golang:1.21-alpine AS builder

COPY . /app

WORKDIR /app

RUN go build -o gnofaucet ./cmd


FROM alpine

COPY --from=builder /app/gnofaucet /usr/local/bin/gnofaucet

ENTRYPOINT [ "/usr/local/bin/gnofaucet" ]
