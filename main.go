package main

import (
	"flag"
	"log"
	"net/http"
	_ "net/http/pprof" // Adds http://*/debug/pprof/ to default mux.
	"runtime"
	"time"
)

var (
	listenAddr      = flag.String("listen", ":5309", "[IP]:port to listen for incoming connections")
	filename        = flag.String("filename", "", "Filename of image to play")
	imageFrameQueue = flag.Int("image-frame-queue", 5, "Image frame queue depth")
	baudRate        = flag.Int("baud-rate", 115200, "Baud rate of serial port")
	numPixels       = flag.Int("num-pixels", 2448, "Number of pixels on USB controller")
	serialPort      = flag.String("serial-port", "", "Serial port to open")
	frameDelay      = flag.Duration("frame-delay", time.Second/30, "Delay between sending frames")
	brightness      = flag.Int("brightness", 255, "Brightness value of LEDs (max 255)")
	imagePath       = flag.String("image-path", "", "Directory of images to play")
)

func main() {
	flag.Parse()

	runtime.GOMAXPROCS(runtime.NumCPU())

	if *serialPort == "" {
		log.Fatal("-serial-port must be set")
	}

	if *numPixels <= 0 || *numPixels > 10000 {
		log.Fatal("-num-pixels must be > 0 and <= 10000")
	}

	if *frameDelay < time.Second/1000 {
		log.Fatal("-frameDelay must be > 0.001s")
	}

	if *brightness <= 0 || *brightness > 255 {
		log.Fatal("-brightness must be > 0 and <= 255")
	}

	sender := Sender{
		SerialPort: *serialPort,
		BaudRate:   *baudRate,
		NumPixels:  *numPixels * 3,
		Brightness: 255,
	}
	streamer := NewStreamer()
	sc := make(chan Frame, *imageFrameQueue)
	go sender.Worker(sc)
	go streamer.Worker(sc, *frameDelay)

	decoder := NewDecoder(*imagePath)
	if decoder == nil {
		log.Fatal("-image-path contains no valid images")
	}
	streamer.SetFramer(decoder)

	log.Fatal(http.ListenAndServe(*listenAddr, nil))
}
