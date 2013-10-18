package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"

	"github.com/Lavos/gofocus"
)

var (
	config_filename = flag.String("c", "", "filename of json configuration file")
)

func main() {
	flag.Parse()

	// configuration JSON
	config_file, err := os.Open(*config_filename)
	defer config_file.Close()

	if err != nil {
		log.Fatal(err)
	}

	data := make([]byte, 2048)
	n, err := config_file.Read(data)

	if err != nil {
		log.Fatal(err)
	}

	var config gofocus.Configuration
	err = json.Unmarshal(data[:n], &config)

	if err != nil {
		log.Fatal(err)
	}

	a := gofocus.NewApplication(&config)

	a.Run()
}
