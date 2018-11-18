# Video to JPEG

To get interesting, flowy moving content for the LEDs, I sample pixels out of video content and store them as JPEGs, with the current state of all LEDs represented as the X axis (regardless of orientation) and time as the Y axis.

This converter:

* Uses ffmpeg to decode video to individual frames.
* Loads individual frames in RAM.
* Applies a mask to each frame, sampling pixels in the order they're represented over USB to the controller.
* For each minute of video, generates one 1224x1800 JPEG image, representing the state of 1224 pixels over 1800 frames, and saves it.

## Usage

```
brew install golang ffmpeg youtube-dl
go build .
youtube-dl -f 'bestvideo[ext=mp4]+bestaudio[ext=m4a]/mp4' https://www.youtube.com/watch?v=yI1Wr-mKjT4
./video-to-jpg -video 'Trippy Visual - Marijuana-yI1Wr-mKjT4.mp4' -out-pattern trippy-%04d.jpg
```
