package web

import "embed"

//go:embed index.gohtml
var indexPage embed.FS

func IndexPage() ([]byte, error) {
	return indexPage.ReadFile("index.gohtml")
}
