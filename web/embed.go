package web

import "embed"

//go:embed index.gohtml
var indexPage embed.FS

func IndexPage() (string, error) {
	data, err := indexPage.ReadFile("index.gohtml")
	if err != nil {
		return "", err
	}
	return string(data), nil
}
