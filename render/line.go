package render

import "github.com/BurntSushi/freetype-go/freetype/truetype"

type Chunk struct {
	Text string
	Font *truetype.Font
}

func ChunksFromString(text string, fl FontList) []Chunk {
	var prev *truetype.Font
	chunks := []Chunk{}
	t := ""
	for _, r := range text {
		font := fl.FontSupporting(r)
		if font == prev {
			t += string(r)
		} else {
			chunks = append(chunks, Chunk{t, font})
			t = ""
			prev = font
		}
	}
	// In case everything is in the same font
	if t != "" && prev != nil {
		chunks = append(chunks, Chunk{t, prev})
	}
	return chunks
}
