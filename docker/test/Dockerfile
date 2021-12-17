FROM golang:1.16-alpine

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go test -c -o bin/appTests cmd/app/app.go cmd/app/model.go cmd/app/main.go cmd/app/main_test.go

EXPOSE 3000

CMD ["./bin/appTests", "-test.v"]
