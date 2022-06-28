package main

import (
	"errors"
	"image"
	"strconv"
	"strings"
)

var (
	ErrNoData        = errors.New("no pixel data supplied")
	ErrInvalidOffset = errors.New("invalid pixel offset")
)

type Frame []byte

// ImageRowToFrame copies an image.Image row to a Frame
func ImageRowToFrame(img image.Image, y int) Frame {
	bounds := img.Bounds()

	width := bounds.Max.X - bounds.Min.X
	f := make([]byte, width*3)

	if y >= bounds.Min.Y && y < bounds.Max.Y {
		o := 0
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			f[o] = byte(r >> 8)
			f[o+1] = byte(g >> 8)
			f[o+2] = byte(b >> 8)
			o += 3
		}
	}

	return f
}

func PixelListToFrame(px int, pl string) (Frame, error) {
	f := make([]byte, px*3)
	o := 0
	for _, s := range strings.Split(pl, ",") {
		v, err := strconv.Atoi(strings.TrimSpace(s))
		if err != nil {
			return nil, err
		}
		if v <= 0 || v+o > px {
			return nil, ErrInvalidOffset
		}

		f[o*3+1] = 0xff // Green start
		o += v
		f[(o-1)*3+0] = 0xff // Red end
	}

	return f, nil
}

// Resize makes sure we have exactly num pixels.  If not, repeat existing or
// truncate.
func (a Frame) Resize(num int) (Frame, error) {
	if len(a) == 0 {
		return nil, ErrNoData
	}

	for {
		l := num - len(a)
		switch {
		case l == 0:
			return a, nil
		case l < 0:
			a = a[:num]
			return a, nil
		case l >= len(a):
			l = len(a)
		}
		a = append(a, a[:l]...)
	}
}

func SameSize(a, b Frame) (Frame, Frame, error) {
	l := len(a)
	if l < len(b) {
		l = len(b)
	}
	a, err := a.Resize(l)
	if err != nil {
		return nil, nil, err
	}
	b, err = b.Resize(l)
	if err != nil {
		return nil, nil, err
	}
	return a, b, nil
}

// Scale multiplies each byte by (mult/256)
func (a Frame) Scale(mult int) Frame {
	l := len(a)
	f := make(Frame, l)
	for i := 0; i < l; i++ {
		p := (int(a[i])*mult + 128) / 256
		if p > 255 {
			p = 255
		}
		f[i] = byte(p)
	}

	return f
}

// Add the RGB values of a and b
func (a Frame) Add(b Frame) Frame {
	a, b, err := SameSize(a, b)
	if err != nil {
		return nil
	}

	l := len(a)
	f := make(Frame, l)
	for i := 0; i < l; i++ {
		p := int(a[i]) + int(b[i])
		if p > 255 {
			p = 255
		}
		f[i] = byte(p)
	}

	return f
}

// Merge keeps the brighter of each RGB pixel from a or b
func (a Frame) Merge(b Frame) Frame {
	a, b, err := SameSize(a, b)
	if err != nil {
		return nil
	}

	l := len(a)
	f := make(Frame, l)
	for i := 0; i+2 < l; i += 3 {
		av := int(a[i]) + int(a[i+1]) + int(a[i+2])
		bv := int(b[i]) + int(b[i+1]) + int(b[i+2])
		if av > bv {
			f[i] = a[i]
			f[i+1] = a[i+1]
			f[i+2] = a[i+2]
		} else {
			f[i] = b[i]
			f[i+1] = b[i+1]
			f[i+2] = b[i+2]
		}
	}

	return f
}

// Mult multiplies each pixel of a with b
func (a Frame) Mult(b Frame) Frame {
	a, b, err := SameSize(a, b)
	if err != nil {
		return nil
	}

	l := len(a)
	f := make(Frame, l)
	for i := 0; i < l; i++ {
		f[i] = byte((int(a[i]) * int(b[i])) / 255)
	}

	return f
}

func (a Frame) NextFrame() Frame {
	return a
}

func (a Frame) Close() {
}
