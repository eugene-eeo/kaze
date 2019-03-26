package render

import "github.com/BurntSushi/freetype-go/freetype/truetype"

type FontList []*truetype.Font

func (f FontList) FontSupporting(x rune) *truetype.Font {
	if len(f) == 0 {
		return nil
	}
	// font.Index(...) == 0 if the font doesn't have a glyph for that rune.
	last := f[len(f)-1]
	for _, font := range f {
		if font.Index(x) > 0 {
			return font
		}
	}
	return last
}
