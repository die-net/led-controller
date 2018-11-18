# USB to OctoWS2811

A Teensy 3.2 sketch to receive pixels over the serial port at 12 megabits, buffer them to local memory, and write them to an OctoWS2811 adapter, which drives up to 8 sets of 2811-protocol LED strips.

Pixel frames start with an "*", followed by a byte for brightness (0-255), followed by bytes for the Red, Green, and Blue values for a hardcoded number of pixels (currently 2448).

Frame rate is determined by the sender; how ever often frames are received, they are sent to the LEDs.  Maximum frame rate is limited to 12000000 / (pixels * 3 + 2) or ~204 frames per second for 2448 pixels.

Based on the requested brightness and color values for every pixel, a guess at the total power draw is 

After each frame is received, a JSON-formatted response is sent with the
following fields:

```
{
    "brightness": 0-255 (possibly reduced from requested value by the current limiter),
    "supply_mw": guess at consumed power supply milliwatts consumed by requested pixels,
    audio_mv: {
            count: count of audio samples received,
            min: minimum in millivolts,
            avg: average in millivolts,
            max: maximum in millivolts
    }
}
```

## Usage

Open this Arduino sketch in the Arduino editor app and plug the Teens into the computer with USB.  Press the programming button on the Teensy.
