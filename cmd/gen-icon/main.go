// gen-icon generates assets/icon.png — the gl1tch TDF block-G favicon as a 32×32 PNG.
// Run: go run ./cmd/gen-icon
package main

import (
	"image"
	"image/color"
	"image/png"
	"os"
)

func main() {
	const (
		width  = 32
		height = 32
	)

	// Dracula palette
	bg   := color.RGBA{0x28, 0x2a, 0x36, 0xff} // #282a36
	fill := color.RGBA{0xbd, 0x93, 0xf9, 0xff}  // #bd93f9

	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// Fill background
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, bg)
		}
	}

	// Grid: 8 columns × 4px wide, 5 rows × 5px tall, top offset y=4
	// Each cell: x = col*4, y = 4 + row*5, width=4, height=5
	//
	// Row 0: .XXXXXX.   cols 1-6
	// Row 1: XX......   cols 0-1
	// Row 2: XX..XXX.   cols 0-1, 4-6
	// Row 3: XX...XX.   cols 0-1, 5-6
	// Row 4: .XXXXXX.   cols 1-6

	rows := [][]int{
		{1, 2, 3, 4, 5, 6},    // row 0
		{0, 1},                // row 1
		{0, 1, 4, 5, 6},       // row 2
		{0, 1, 5, 6},          // row 3
		{1, 2, 3, 4, 5, 6},    // row 4
	}

	cellW := 4
	cellH := 5
	topOffset := 4

	for rowIdx, cols := range rows {
		for _, col := range cols {
			x0 := col * cellW
			y0 := topOffset + rowIdx*cellH
			for dy := 0; dy < cellH; dy++ {
				for dx := 0; dx < cellW; dx++ {
					img.Set(x0+dx, y0+dy, fill)
				}
			}
		}
	}

	f, err := os.Create("assets/icon.png")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	if err := png.Encode(f, img); err != nil {
		panic(err)
	}
}
