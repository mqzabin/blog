package extensions

import (
	"github.com/fatih/color"

	"io"
)

type colorWriter struct {
	c *color.Color
	w io.Writer
}

func NewColorWriter(c *color.Color, w io.Writer) io.Writer {
	return &colorWriter{c, w}
}

func (cw *colorWriter) Write(p []byte) (int, error) {
	return cw.c.Fprintf(cw.w, string(p))
}
