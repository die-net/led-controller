package main

import (
	"fmt"
	"os"
	"os/exec"
)

func videoToFrames(video, workDir string) error {
	outFile := workDir + "/" + "frame-%06d.jpg"
	scale := fmt.Sprintf("scale=%d:%d", *videoWidth, *videoHeight)
	args := []string{"-y", "-v", "error", "-i", video, "-threads", "4", "-pix_fmt", "yuv420p", "-sws_flags", "lanczos", "-vf", scale, "-ss", "00:00:00.000", "-f", "image2", outFile}
	return run("ffmpeg", args...)
}

func run(name string, arg ...string) error {
	cmd := exec.Command(name, arg...)
	cmd.Env = os.Environ()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
