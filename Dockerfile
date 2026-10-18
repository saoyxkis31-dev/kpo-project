FROM golang:1.26.6 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -o /bin/service ./cmd/server/

FROM alpine:3.19

COPY --from=builder /bin/service /bin/service

EXPOSE 8080

CMD ["/bin/service"]