# Radio Stream Server

Это простой Go-сервер, который сканирует указанные папки с музыкальными файлами (`.mp3`, `.aac`, `.ogg`), перекодирует их в заданный формат и битрейт при запуске (например, в `mp3@96kbps`), и раздаёт как потоковое радио по HTTP. Поддерживает несколько станций, одну на папку, единый поток для всех клиентов (live-режим).

## 🚀 Возможности

- Стриминг через HTTP по TCP: `/stream/:station`
- `/stations` — список всех станций (папок)
- `/nowplaying/:station` — текущий трек
- `/` — простая стартовая страница
- Буферизация (по умолчанию 7 секунд)
- Предварительная перекодировка треков через `ffmpeg`
- Использует только один поток на станцию (экономия ресурсов)

## ⚙️ Конфигурация

### Порядок загрузки:
1. `config.yaml` — загружается при старте
2. ENV-переменные перезаписывают значения из `config.yaml`

### Пример `config.yaml`

```yaml
ffmpeg_path: /usr/bin/ffmpeg
output_format: mp3
output_bitrate: 96k
music_dir: ./music
cache_dir: ./.cache
buffer_seconds: 7
listen: ":8080"
```

### ENV переменные (все необязательные)

| Переменная | Значение |
|-----------|----------|
| `FFMPEG_PATH` | Путь до ffmpeg |
| `OUTPUT_FORMAT` | Формат выходных файлов (`mp3`, `aac`, `ogg`) |
| `OUTPUT_BITRATE` | Например `96k` |
| `MUSIC_DIR` | Папка с исходными треками |
| `CACHE_DIR` | Папка с перекодированными |
| `BUFFER_SECONDS` | Кол-во секунд для предварительной буферизации |
| `LISTEN` | Порт и IP (по умолчанию `:8080`) |

## 🐳 Docker (для MikroTik)

```Dockerfile
FROM golang:1.22-alpine

RUN apk add --no-cache ffmpeg
WORKDIR /app
COPY . .

RUN go build -o radiod ./cmd/radiod

EXPOSE 8080
CMD ["./radiod"]
```

## 🗂 Структура проекта

```
cmd/radiod         # main
internal/server    # HTTP роуты
internal/stream    # логика проигрывания
internal/convert   # ffmpeg обёртка
internal/config    # загрузка конфигов
web/               # статичные файлы (если надо)
```

## 🔗 Эндпоинты

- `/` — HTML страница со списком станций
- `/stations` — JSON массив: `["pop", "rock", "electro"]`
- `/nowplaying/:station` — JSON: `{ "now": "track1.mp3" }`
- `/stream/:station` — HTTP-поток `audio/mpeg`/...

