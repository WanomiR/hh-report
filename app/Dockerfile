FROM golang:1.22-alpine as builder

WORKDIR /app

COPY . .

RUN go mod init app && go mod tidy

RUN go build -o main ./cmd/app/

FROM alpine:latest

COPY --from=builder /app/main /main
COPY --from=builder /app/.env /.env

EXPOSE 1569

CMD ["/main"]
