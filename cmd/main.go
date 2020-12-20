package main

import (
	"github.com/eriksywu/ascii/cmd/server"
	"github.com/eriksywu/ascii/pkg/filestore"
	"github.com/eriksywu/ascii/pkg/image"
	"log"
)

const defaultStorePath = "/asciistore"

//TODO set this as a flag
//we can mount a hostvolume or pvc to persist
var StorePath = defaultStorePath

func main() {
	imageStore, err := filestore.NewStore(StorePath)
	if err != nil {
		log.Fatal(err)
	}
	asciiService := image.NewService(imageStore)
	app := server.BuildServer(asciiService, 8000)
	app.Run()
}
