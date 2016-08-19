package main

import (
	"errors"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

var (
	ErrUnknownFormat = errors.New("Image is in unknown format")
)

type Decoder struct {
	files   []string
	fileNum int
	image   image.Image
	y       int
}

func NewDecoder(path string) *Decoder {
	files, err := getFilenames(path)
	if err != nil {
		return nil
	}

	d := &Decoder{
		files:   files,
		fileNum: 0,
		image:   nil,
		y:       0,
	}

	if d.NextImage() {
		return d
	}

	return nil
}

func getFilenames(path string) ([]string, error) {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return []string{}, err
	}

	o := []string{}
	for _, file := range files {
		n := file.Name()
		if strings.HasPrefix(n, ".") || file.IsDir() {
			continue
		}
		o = append(o, path+"/"+n)
	}

	return o, nil
}

func (d *Decoder) NextImage() bool {
	d.image = nil
	d.y = 0
	for {
		if len(d.files) == 0 {
			return false
		}

		if d.fileNum > len(d.files) {
			d.fileNum = 0
		}
		file := d.files[d.fileNum]

		img, err := readImage(file)
		if err == nil {
			d.image = img
			d.y = img.Bounds().Min.Y
			d.fileNum++
			return true
		}

		log.Println("Error reading", file, err)
		d.files = append(d.files[:d.fileNum], d.files[d.fileNum+1:]...)
	}
}

func (d *Decoder) NextFrame() Frame {
	if d.image == nil {
		return nil
	}

	f := ImageRowToFrame(d.image, d.y)

	// If we're on the last row, we'll have to load the next image.
	d.y++
	if d.y > d.image.Bounds().Max.Y {
		d.NextImage()
	}

	return f
}

func readImage(file string) (image.Image, error) {
	reader, err := os.Open(file)
	if err != nil {
		return nil, err
	}

	defer reader.Close()

	img, _, err := image.Decode(reader)
	if err != nil {
		return nil, err
	}

	return img, nil
}

func (d *Decoder) Close() {
}