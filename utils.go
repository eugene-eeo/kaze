package main

import "os"
import "image"
import "github.com/BurntSushi/xgbutil"
import "github.com/BurntSushi/xgbutil/xgraphics"
import "github.com/BurntSushi/freetype-go/freetype/truetype"

func maxWidth(text string, max int, oracle func(string) int) string {
	n := len(text)
	l := oracle(text)
	i := n
	// minimisation stage
	for l > max && i > 0 {
		i /= 2
		l = oracle(text[:i])
	}
	// maximisation stage
	for i < n {
		i++
		l = oracle(text[:i])
		if l > max {
			i--
			break
		}
	}
	return text[:i]
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func ximgWithProps(X *xgbutil.XUtil, padding, height, width, border int, bg, borderColor xgraphics.BGRA) *xgraphics.Image {
	width += 2*padding + 2*border
	height += 2*padding + 2*border
	ximg := xgraphics.New(X, image.Rect(0, 0, width, height))
	ximg.For(func(x, y int) xgraphics.BGRA {
		// top, left, right, bottom borders
		if x < border || y < border || x >= width-border || y >= height-border {
			return borderColor
		}
		return bg
	})
	return ximg
}

func mustReadFont(path string) *truetype.Font {
	fontReader, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	font, err := xgraphics.ParseFont(fontReader)
	return xgraphics.MustFont(font, err)
}
