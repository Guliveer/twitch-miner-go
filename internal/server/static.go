package server

import (
	"embed"
	"io/fs"

	"github.com/Guliveer/twitch-miner-go/internal/feedback"
)

//go:embed static
var staticEmbed embed.FS

var staticFS fs.FS

var (
	dashboardHTML []byte
	logsHTML      []byte
)

func init() {
	var err error
	staticFS, err = fs.Sub(staticEmbed, "static")
	if err != nil {
		panic("server: failed to create static sub-filesystem: " + err.Error())
	}

	dashboardHTML, err = loadPage("static/index.html")
	if err != nil {
		panic("server: failed to read embedded dashboard HTML: " + err.Error())
	}

	logsHTML, err = loadPage("static/logs.html")
	if err != nil {
		panic("server: failed to read embedded logs HTML: " + err.Error())
	}
}

// loadPage reads an embedded page and fills in the version and feedback links
// so the shipped HTML stays a plain static file.
func loadPage(name string) ([]byte, error) {
	page, err := staticEmbed.ReadFile(name)
	if err != nil {
		return nil, err
	}
	return []byte(feedback.Expand(string(page))), nil
}
