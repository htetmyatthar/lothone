package static

import (
	"embed"
)

//go:embed static/css/* static/js/* static/images/*
var WebFS embed.FS
