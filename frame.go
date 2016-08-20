package main

import (
	"errors"
	"image"
)

var (
	ErrNoData = errors.New("No pixel data supplied")
)

type Frame []byte

// Copy an image.Image row to a Frame
func ImageRowToFrame(img image.Image, y int) Frame {
	bounds := img.Bounds()

	width := bounds.Max.X - bounds.Min.X
	f := make([]byte, width*3, width*3)

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

// Make sure we have exactly num pixels.  If not, repeat existing or
// truncate.
func (f Frame) Resize(num int) (Frame, error) {
	if f == nil || len(f) == 0 {
		return nil, ErrNoData
	}

	for {
		l := num - len(f)
		switch {
		case l == 0:
			return f, nil
		case l < 0:
			f = f[:num]
			return f, nil
		case l >= len(f):
			l = len(f)
		}
		f = append(f, f[:l]...)
	}
}

func SameSize(a Frame, b Frame) (Frame, Frame, error) {
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
	f := make(Frame, l, l)
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
	f := make(Frame, l, l)
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
	f := make(Frame, l, l)
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
	f := make(Frame, l, l)
	for i := 0; i < l; i++ {
		f[i] = byte((int(a[i]) * int(b[i])) / 255)
	}

	return f
}

func (f Frame) NextFrame() Frame {
	return f
}

func (f Frame) Close() {
}
