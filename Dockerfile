# Используем официальный минималистичный образ Golang
FROM golang:1.22-alpine

# Добавляем ffmpeg для перекодировки аудио
RUN apk add --no-cache ffmpeg

# Рабочая директория для приложения
WORKDIR /app

# Копируем исходники
COPY . .

# Сборка бинарника радиосервера
RUN go build -o radiod ./cmd/radiod

# Публикуем порт HTTP
EXPOSE 8080

# Запускаем сервер
CMD ["./radiod"]
