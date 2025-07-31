package server

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/yosssi/gohtml"
	"x07-it/radiod/internal/stream"
)

// SetupRouter creates and configures Gin router with all endpoints.
func SetupRouter(p *stream.Player) *gin.Engine {
	// Загружаем шаблон главной страницы. В случае ошибки используем запасной HTML.
	tmplBytes, err := os.ReadFile("web/index.gohtml")
	var indexTemplate string
	if err != nil {
		logrus.WithError(err).Warn("fallback to built-in index template")
		indexTemplate = "<html><body><h1>Stations</h1><ul>{{STATIONS}}</ul></body></html>"
	} else {
		indexTemplate = string(tmplBytes)
	}

	r := gin.Default()

	// Index page with list of stations.
	r.GET("/", func(c *gin.Context) {
		stations := p.StationNames()
		var b strings.Builder
		for _, s := range stations {
			b.WriteString(fmt.Sprintf("<li><a href=\"/stream/%s\">%s</a></li>", s, s))
		}
		// Вставляем список станций в шаблон.
		page := strings.Replace(indexTemplate, "{{STATIONS}}", b.String(), 1)
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
		defer remove()
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
