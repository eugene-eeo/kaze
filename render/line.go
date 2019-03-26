package render

import "github.com/BurntSushi/xgbutil/xgraphics"
import "github.com/BurntSushi/freetype-go/freetype/truetype"

type Chunk struct {
	W    int
	H    int
	Rune rune
	Font *truetype.Font
}

type Line struct {
	Chunks []Chunk
	Height int
	Width  int
}

type TextBox []Line

func ChunksFromString(text string, fl FontList, fontSize float64, cb func(Chunk)) {
	c := Chunk{}
	for _, r := range text {
		font := fl.FontSupporting(r)
		c.W, c.H = xgraphics.Extents(font, fontSize, string(r))
		c.Font = font
		c.Rune = r
		cb(c)
	}
}

func TextBoxWithMaxWidth(text string, fl FontList, fontSize float64, maxWidth int) TextBox {
	tb := TextBox{}
	line := Line{Chunks: []Chunk{}}
	ChunksFromString(text, fl, fontSize, func(c Chunk) {
		if line.Width+c.W > maxWidth {
			tb = append(tb, line)
			line = Line{
				Chunks: []Chunk{c},
				Height: c.H,
				Width:  c.W,
			}
		} else {
			line.Chunks = append(line.Chunks, c)
			line.Width += c.W
			if line.Height < c.H {
				line.Height = c.H
			}
		}
	})
	return tb
}
