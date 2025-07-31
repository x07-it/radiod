package server

import (
	"fmt"
	"io"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/yosssi/gohtml"
	"x07-it/radiod/internal/stream"
)

// SetupRouter creates and configures Gin router with all endpoints.
func SetupRouter(p *stream.Player) *gin.Engine {
	r := gin.Default()

	// Index page with list of stations.
	r.GET("/", func(c *gin.Context) {
		stations := p.StationNames()
		var b strings.Builder
		b.WriteString("<html><body><h1>Stations</h1><ul>")
		for _, s := range stations {
			b.WriteString(fmt.Sprintf("<li><a href=\"/stream/%s\">%s</a></li>", s, s))
		}
		b.WriteString("</ul></body></html>")
		html := gohtml.Format(b.String())
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
