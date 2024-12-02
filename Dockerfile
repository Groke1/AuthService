FROM golang:1.23

WORKDIR /src/auth_service

COPY . .

RUN go mod tidy

EXPOSE 8080

ENTRYPOINT ["go", "run", "cmd/main.go"]