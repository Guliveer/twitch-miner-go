package configeditor

import "embed"

//go:embed web/index.html web/app.js web/style.css
var webFS embed.FS
