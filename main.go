package main

import (
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var quality *int
var jxlTransOnly *int

func main() {
	// Get directory
	// convert every image that's .jpeg/.jpg/.png into the following
	// avif
	// jxl
	directory := flag.String("dir", "", "directory containing images (required)")
	quality = flag.Int("quality", 75, "quality argument to pass to libraries")
	jxlTransOnly = flag.Int("lossless_jpeg", 1, "Only transcode jpeg to jxl, don't reprocess (default: true)")
	flag.Parse()
	if len(*directory) <= 0 {
		fmt.Fprintf(os.Stderr, "dir is a required flag\n")
		os.Exit(1)
	}
	files, err := os.ReadDir(*directory)

	if err != nil {
		log.Fatal(err)
	}
	// check if libraries are installed
	_, err = exec.LookPath("avifenc")
	if err != nil {
		log.Fatal("Could not find libavif installation. Check to see if you've installed the library.")
	}
	_, err = exec.LookPath("cjxl")
	if err != nil {
		log.Fatal("Could not find libjxl installation. Check to see if you've installed the library.")
	}

	// TODO: Consider parallelizing this.
	for _, filename := range files {
		fmt.Println(filename.Name())
		fileExtOnly := filepath.Ext(filename.Name())
		fmt.Println(fileExtOnly)
		switch fileExtOnly {
		case ".jpeg", ".jpg":
			convertToAVIF(filename, *directory)
			convertToJXL(filename, *directory)
		default:
			fmt.Println("Unrecognized File Extension. Skipping.")
		}
	}
}

// Converts input to AVIF file at quality specified on the command line
func convertToAVIF(filename fs.DirEntry, directory string) {
	fileNameOnly := strings.TrimSuffix(filepath.Base(filename.Name()), filepath.Ext(filename.Name()))
	fileInfo, err := os.Stat(fmt.Sprintf("%s/%s.avif", directory, fileNameOnly))
	if err == nil && !fileInfo.Mode().IsRegular() {
		fmt.Println("File exists. Breaking...")
	}
	cmd := exec.Command(
		"avifenc",
		"-q", fmt.Sprintf("%d", *quality),
		"--min", "0",
		"--max", "60",
		"-a", "end-usage=q",
		"-a", "cq-level=18",
		"-a", "tune=ssim",
		"--jobs", "8",
		filepath.Join(directory, filename.Name()),
		fmt.Sprintf("%s/%s.avif", directory, fileNameOnly),
	)
	err = cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
}

// Convert input filename from whatever it is into a JPEG XL :)
func convertToJXL(filename fs.DirEntry, directory string) {
	fileNameOnly := strings.TrimSuffix(filepath.Base(filename.Name()), filepath.Ext(filename.Name()))
	fileInfo, err := os.Stat(fmt.Sprintf("%s/%s.jxl", directory, fileNameOnly))
	if err == nil && !fileInfo.Mode().IsRegular() {
		fmt.Println("File exists. Breaking...")
	}
	jxlQuality := *quality
	if *jxlTransOnly == 1 {
		jxlQuality = 100
	}
	cmd := exec.Command(
		"cjxl",
		"-q", fmt.Sprintf("%d", jxlQuality),
		"--num_threads", "-1",
		"-j", fmt.Sprintf("%d", *jxlTransOnly),
		filepath.Join(directory, filename.Name()),
		fmt.Sprintf("%s/%s.jxl", directory, fileNameOnly),
	)
	err = cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
}
