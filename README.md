# LED Controller [![Build Status](https://github.com/die-net/led-controller/actions/workflows/go-test.yml/badge.svg)](https://github.com/die-net/led-controller/actions/workflows/go-test.yml) [![Go Report Card](https://goreportcard.com/badge/github.com/die-net/led-controller)](https://goreportcard.com/report/github.com/die-net/led-controller)

For my Burning Man 2016 art car, I need to do something interesting with a few thousand individually addressable 2813 + 5050 LEDs.

I ran across the [whisp demo video](https://www.youtube.com/watch?v=UQoA1foNbnQ) of KentuckyFriedFrank's [pngled_server](https://github.com/KentuckyFriedFrank/pngled_server) and was quite enamoured of the result.

It uses a combination of a [Raspberry Pi 3](https://www.raspberrypi.org/products/raspberry-pi-3-model-b/), a [Teensy 3.2](https://www.pjrc.com/store/teensy32.html), and [OctoWS2811](https://www.pjrc.com/store/octo28_adaptor.html) break out board to control up to 8 different channels of 2811/2812-protocol addressable LEDs.

I started with similar hardware, and decided to rewrite the Raspberry Pi controller from Python to Go, and substantially refactor the Teensy controller for my needs.

This unlikely to be directly useful to anyone else, though feel free to borrow heavily and/or customize for your needs.
