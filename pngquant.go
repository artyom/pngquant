// Command pngquant reduces number of colors in a given PNG image to a fixed
// palette.
package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"

	"github.com/artyom/autoflags"
	"github.com/soniakeys/quant/median"
)

func main() {
	args := struct {
		In  string `flag:"in,input file"`
		Out string `flag:"out,output file"`
		N   int    `flag:"n,number of colors (up to 256)"`
		ND  bool   `flag:"nodither,do not apply dithering"`
		T   bool   `flag:"trans,keep transparency if present"`
	}{N: 256}
	autoflags.Parse(&args)
	if err := do(args.Out, args.In, args.N, !args.ND, args.T); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func do(outName, inName string, n int, dither, keepTrans bool) error {
	if n <= 0 || n > 256 {
		return fmt.Errorf("unsupported number of colors: %d", n)
	}
	if outName == "" || inName == "" || outName == inName {
		return fmt.Errorf("both input and output names should be set to different non-empty values")
	}
	f, err := os.Open(inName)
	if err != nil {
		return err
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	if err != nil {
		return err
	}
	var withTransparency bool
	if op, ok := img.(interface{ Opaque() bool }); keepTrans && ok {
		withTransparency = !op.Opaque()
	}
	out, err := os.Create(outName)
	if err != nil {
		return err
	}
	defer out.Close()
	pal := make(color.Palette, 0, n)
	if withTransparency {
		pal = append(pal, color.Transparent)
	}
	pal = median.Quantizer(n).Quantize(pal, img)
	b := img.Bounds()
	imgp := image.NewPaletted(b, pal)
	switch {
	case dither:
		draw.FloydSteinberg.Draw(imgp, b, img, b.Min)
	default:
		draw.Draw(imgp, b, img, b.Min, draw.Over)
	}
	if err := (&png.Encoder{CompressionLevel: png.BestCompression}).Encode(out, imgp); err != nil {
		os.Remove(out.Name())
		return err
	}
	return out.Close()
}
