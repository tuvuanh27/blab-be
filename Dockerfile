FROM golang:latest

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .
COPY .env /app

RUN go build -o main .

EXPOSE 8080

CMD ["./main"]
