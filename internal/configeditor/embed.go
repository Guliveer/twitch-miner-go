package configeditor

import (
	"embed"

	"github.com/Guliveer/twitch-miner-go/internal/feedback"
)

//go:embed web/index.html web/app.js web/style.css
var webFS embed.FS

// indexHTML is the editor shell with the version and feedback links filled in.
var indexHTML = func() []byte {
	page, err := webFS.ReadFile("web/index.html")
	if err != nil {
		panic("configeditor: failed to read embedded index.html: " + err.Error())
	}
	return []byte(feedback.Expand(string(page)))
}()
