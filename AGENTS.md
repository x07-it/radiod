# AGENTS.md

## 🎯 Цель

Создать HTTP-аудио-сервер на языке Go, который:

1. При старте сканирует указанную директорию с музыкальными файлами (`.mp3`, `.aac`, `.ogg`)
2. Перекодирует каждый файл в `output_format` и `output_bitrate` (используя `ffmpeg`)
3. Сохраняет перекодированные версии в кэше (`cache_dir`)
4. Запускает `stream loop` для каждой папки (станции)
5. Отдаёт поток по `/stream/:station`, синхронно для всех клиентов
6. Предоставляет эндпоинты `/stations`, `/nowplaying/:station`, `/`

## 📦 Библиотеки

- `github.com/gin-gonic/gin` — HTTP сервер
- `github.com/sirupsen/logrus` — логирование
- `github.com/yosssi/gohtml` — генерация HTML
- `github.com/spf13/viper` — конфиг и ENV
- `os/exec` — вызов `ffmpeg`
- `bufio`, `os`, `io` — работа с файлами и буферами

## 🧱 Архитектура

- `cmd/radiod/main.go` — старт
- `internal/config/config.go` — загрузка конфигурации
- `internal/convert/convert.go` — логика перекодировки через `ffmpeg`
- `internal/stream/player.go` — стриминг и очередь воспроизведения
- `internal/server/routes.go` — маршруты Gin
- `web/index.gohtml` — простая страница с `gohtml`

## 📁 Эндпоинты

| Route | Method | Description |
|-------|--------|-------------|
| `/` | GET | HTML со списком станций |
| `/stations` | GET | JSON список станций |
| `/nowplaying/:station` | GET | JSON с текущим треком |
| `/stream/:station` | GET | HTTP-поток mp3 |

## 🔧 Параметры запуска

Загружаются из `config.yaml`, затем из ENV. См. `README.md`.

## 🐳 Docker инструкция

Создай `Dockerfile` с:

```Dockerfile
FROM golang:1.22-alpine
RUN apk add --no-cache ffmpeg
WORKDIR /app
COPY . .
RUN go build -o radiod ./cmd/radiod
EXPOSE 8080
CMD ["./radiod"]
```

## 🧠 Особенности

- Стриминг синхронный: все клиенты получают одинаковый поток
- Буферизация реализуется через `bufio` с таймером
- Перекодирование происходит только один раз при запуске
- Если файл уже перекодирован — повторно не обрабатывается

## ❌ Что не реализуется

- Поддержка ICY/SHOUTcast
- WebSocket
- Метаданные в потоке
- Персонализированный буфер/seek

## ✅ Проверка

1. Помести `.mp3` в `music/pop`, `music/rock`
2. Запусти сервер
3. Перейди на `http://localhost:8080/stream/pop` — начнётся воспроизведение
4. Проверь `http://localhost:8080/stations` и `nowplaying/pop`

