#docker build -t go-payment-authorizer .
#docker run -dit --name go-payment-authorizer -p 65000:65000 go-payment-authorizer sleep infinity

FROM golang:1.23.3 As builder

RUN apt-get update && apt-get install bash && apt-get install -y --no-install-recommends ca-certificates

WORKDIR /app
COPY . .
RUN go mod tidy

WORKDIR /app/cmd
RUN go build -o go-payment-authorizer -ldflags '-linkmode external -w -extldflags "-static"'

FROM alpine

WORKDIR /app
COPY --from=builder /app/cmd/go-payment-authorizer .
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

CMD ["/app/go-payment-authorizer"]