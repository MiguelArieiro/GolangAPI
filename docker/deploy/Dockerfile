FROM golang:1.16-alpine

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o bin/app cmd/app/app.go cmd/app/model.go cmd/app/main.go cmd/app/main_test.go

EXPOSE 3000

CMD ["./bin/app"]
