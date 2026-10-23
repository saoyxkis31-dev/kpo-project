FROM golang:1.26-alpine AS builder
WORKDIR /app

# копируем список зависимостей и скачиваем их
COPY go.mod go.sum ./
RUN go mod download

# копируем весь остальной код
COPY . .

# компилируем статический бинарник без C-зависимостей
RUN CGO_ENABLED=0 go build -o /bin/service ./cmd/server/main.go

# легковесный образ
FROM alpine:3.19
WORKDIR /app

# переносим готовый бинарник из builder
COPY --from=builder /bin/service /bin/service

EXPOSE 8080

CMD ["/bin/service"]