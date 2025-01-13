FROM golang:1.21.5

WORKDIR /mytheresa

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o main .

CMD ["./main"]
