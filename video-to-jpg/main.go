package main

import (
	"bytes"
	"flag"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/png"
	"io/ioutil"
	"log"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
)

var (
	video          = flag.String("video", "", "Video file to convert")
	outPattern     = flag.String("out-pattern", "out-%04d.jpg", "Output filename path pattern")
	framesPerImage = flag.Int("frames-per-image", 1800, "Maximum number of frames per output image")
	workers        = flag.Int("workers", runtime.NumCPU(), "Image decoding worker threads")
	videoWidth     = flag.Int("video-width", 256, "Scaled video width")
	videoHeight    = flag.Int("video-height", 144, "Scaled video height")
	pixelBoxes     = flag.String("pixelBoxes", "143x114,103x82,67x52,30x21", "Comma seperated list of concentric boxes for pixel path")
	pixels         = []image.Point{}
	rowChan        chan row
	wg             sync.WaitGroup
)

type row struct {
	img      *image.RGBA
	y        int
	filename string
}

func main() {
	flag.Parse()

	if *video == "" {
		log.Fatal("-video must be set.")
	}

	for _, boxStr := range strings.Split(*pixelBoxes, ",") {
		ws, hs := splitTwo(boxStr, "x")
		w, _ := strconv.Atoi(ws)
		h, _ := strconv.Atoi(hs)
		ox := (*videoWidth - w) / 2
		oy := (*videoHeight - h) / 2
		if w < 2 || h < 2 || ox <= 0 || oy <= 0 {
			log.Fatal("width and height must be >= 2 and >= -video-width and -video-height")
		}
		for y := oy + h; y >= oy; y-- {
			pixels = append(pixels, image.Point{ox, y})
		}
		for x := ox; x <= ox+w; x++ {
			pixels = append(pixels, image.Point{x, oy})
		}
		for y := oy; y <= oy+h; y++ {
			pixels = append(pixels, image.Point{ox + w, y})
		}
		for x := ox + w; x >= ox; x-- {
			pixels = append(pixels, image.Point{x, oy + h})
		}
	}

	if len(pixels) <= 0 {
		log.Fatal("-pixelBoxes must be set.")
	}

	runtime.GOMAXPROCS(runtime.NumCPU())

	workDir, err := ioutil.TempDir("", "video-to-jpg")
	if err != nil {
		log.Fatal(err)
	}
	workDir += "/"
	defer os.RemoveAll(workDir)

	err = videoToFrames(*video, workDir)
	if err != nil {
		log.Fatal(err)
	}

	files, err := ioutil.ReadDir(workDir)
	if err != nil {
		log.Fatal(err)
	}

	for count := 1; ; count++ {
		frames := len(files)
		if frames > *framesPerImage {
			frames = *framesPerImage
		}
		if frames <= 0 {
			break
		}

		f := files[:frames]
		files = files[frames:]

		img := image.NewRGBA(image.Rect(0, 0, len(pixels), frames))

		rowChan = make(chan row, 100)

		wg.Add(*workers)
		for i := 0; i < *workers; i++ {
			go rowWorker()
		}

		for y, file := range f {
			rowChan <- row{img: img, y: y, filename: workDir + file.Name()}
		}

		close(rowChan)

		wg.Wait()

		err = writeJpeg(img, fmt.Sprintf(*outPattern, count))
		if err != nil {
			log.Fatal(err)
		}
	}
}

func rowWorker() {
	for row := range rowChan {
		src, err := readImage(row.filename)
		if err != nil {
			log.Fatal(err)
		}

		for x, pixel := range pixels {
			row.img.Set(x, row.y, src.At(pixel.X, pixel.Y))
		}
	}
	wg.Done()
}

func readImage(filename string) (image.Image, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	return img, err
}

func writeJpeg(img image.Image, filename string) error {
	var buf bytes.Buffer
	err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 75})
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(filename, buf.Bytes(), 0644)
	if err != nil {
		return err
	}

	return nil
}

// Simplified strings.SplitN() that always returns two strings.
func splitTwo(s, sep string) (one, two string) {
	if part := strings.SplitN(s, sep, 2); len(part) == 2 {
		return part[0], part[1]
	}

	return s, ""
}
