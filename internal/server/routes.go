package server

import (
	"html/template"
	"io"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/yosssi/gohtml"
	"x07-it/radiod/internal/stream"
)

// SetupRouter creates and configures Gin router with all endpoints.
func SetupRouter(p *stream.Player) *gin.Engine {
	// Шаблон главной страницы хранится в бинарнике.
	indexHTML := string(indexTemplate)
	if len(indexHTML) == 0 {
		logrus.Warn("fallback to built-in index template")
		indexHTML = "<html><body><h1>Stations</h1><ul>{{STATIONS}}</ul></body></html>"
	}

	r := gin.Default()

	// Шаблон ссылки на станцию: `<li><a href="/stream/{{.}}">{{.}}</a></li>`.
	linkTmpl := template.Must(template.New("stationLink").Parse(`<li><a href="/stream/{{.}}">{{.}}</a></li>`))

	// Index page with list of stations.
	r.GET("/", func(c *gin.Context) {
		stations := p.StationNames()
		var b strings.Builder
		for _, s := range stations {
			// Рендерим безопасную ссылку на станцию.
			if err := linkTmpl.Execute(&b, s); err != nil {
				logrus.WithError(err).Error("cannot render station link")
			}
		}
		// Вставляем список станций в шаблон.
		page := strings.Replace(indexHTML, "{{STATIONS}}", b.String(), 1)
		html := gohtml.Format(page)
		c.Data(200, "text/html; charset=utf-8", []byte(html))
	})

	// Return JSON array of available stations.
	r.GET("/stations", func(c *gin.Context) {
		c.JSON(200, p.StationNames())
	})

	// Current playing track for station.
	r.GET("/nowplaying/:station", func(c *gin.Context) {
		st := p.Get(c.Param("station"))
		if st == nil {
			c.JSON(404, gin.H{"error": "station not found"})
			return
		}
		c.JSON(200, gin.H{"now": st.NowPlaying()})
	})

	// Stream endpoint providing audio data to clients.
	r.GET("/stream/:station", func(c *gin.Context) {
		st := p.Get(c.Param("station"))
		if st == nil {
			c.JSON(404, gin.H{"error": "station not found"})
			return
		}
		ch, remove := st.AddListener()

		// Логируем подключение слушателя для аудита и отладки.
		remoteIP := c.ClientIP()
		stationName := c.Param("station")
		logrus.WithFields(logrus.Fields{
			"remote":  remoteIP,
			"station": stationName,
		}).Info("listener connected")

		defer func() {
			remove()
			// При отключении фиксируем событие с теми же данными.
			logrus.WithFields(logrus.Fields{
				"remote":  remoteIP,
				"station": stationName,
			}).Info("listener disconnected")
		}()

		c.Header("Content-Type", "audio/mpeg")
		c.Status(200)
		c.Stream(func(w io.Writer) bool {
			if data, ok := <-ch; ok {
				_, _ = w.Write(data)
				return true
			}
			return false
		})
	})

	return r
}
