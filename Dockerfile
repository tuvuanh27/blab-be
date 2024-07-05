FROM golang:1.22 as builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

# Build the binary statically
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

FROM alpine:3.14 as production

WORKDIR /app

# Create a user 'go' and set permissions
RUN addgroup -S go && adduser -S go -G go

COPY --chown=go:go --from=builder /app/main /app/main

USER go
COPY .env .env

CMD ["./main"]
