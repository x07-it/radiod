package server

import _ "embed"

// indexTemplate содержит шаблон главной страницы, встроенный в бинарник.
//
//go:embed ../../web/index.gohtml
var indexTemplate []byte
