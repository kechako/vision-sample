package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"

	"github.com/pkg/errors"

	vision "cloud.google.com/go/vision/apiv1"
	pb "google.golang.org/genproto/googleapis/cloud/vision/v1"
)

func main() {
	code, err := run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error : %#v\n", err)
	}
	if code != 0 {
		os.Exit(code)
	}
}

func run() (int, error) {
	var count int
	flag.IntVar(&count, "c", 10, "Max results count")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "%s [options] image\n\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	if flag.NArg() != 1 {
		flag.Usage()
		return 1, nil
	}

	var data []byte
	if u, err := url.Parse(flag.Arg(0)); err == nil && (u.Scheme == "http" || u.Scheme == "https") {
		var err error
		data, err = getImage(u.String())
		if err != nil {
			return 1, err
		}
	} else {
		var err error
		data, err = openImage(flag.Arg(0))
		if err != nil {
			return 1, err
		}
	}

	ctx := context.Background()

	client, err := vision.NewImageAnnotatorClient(ctx)
	if err != nil {
		return 1, errors.Wrap(err, "could not create new client")
	}

	image := &pb.Image{
		Content: data,
	}

	labels, err := client.DetectLabels(ctx, image, nil, count)
	if err != nil {
		return 1, errors.Wrap(err, "could not detect labels")
	}

	for _, label := range labels {
		fmt.Printf("Score : %.6f, Description: %s\n", label.Score, label.Description)
	}

	return 0, nil
}

func getImage(url string) ([]byte, error) {
	res, err := http.Get(url)
	if err != nil {
		return nil, errors.Wrapf(err, "could not get image from %s", url)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, errors.Wrapf(err, "server response %s", res.Status)
	}

	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, errors.Wrapf(err, "could not read image data from %s", url)
	}

	return data, nil
}

func openImage(name string) ([]byte, error) {
	file, err := os.Open(name)
	if err != nil {
		return nil, errors.Wrapf(err, "could not open file %s", name)
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, errors.Wrapf(err, "could not read file %s", name)
	}

	return data, nil
}
