package main

import (
	"log"
	"os"
	"io"
	"flag"
	"encoding/json"
	"runtime"
	//"net/http"
	//"bytes"

	"github.com/BurntSushi/toml"
	"github.com/sidepelican/goprobe/probe"
)

type Config struct {
	ApiUrl string
	ApName string
}

func loadConfig(path string) (config Config) {
	if path == "" {
		if runtime.GOOS == "linux" {
			path = "/etc/goprobe/config.tml"
		}
	}

	// decode const settings
	if _, err := toml.DecodeFile(path, &config); err != nil {
		log.Println(err)
		config.ApName = "undefined"
	}
	return
}

func main() {
	// logging setup
	if runtime.GOOS == "linux" {
		f, err := os.OpenFile("/var/log/goprobe.log", os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
		if err != nil {
			println("error opening file: %v", err)
		}
		defer f.Close()
		log.SetOutput(io.MultiWriter(f, os.Stdout)) // assign it to the standard logger
	}
	log.SetFlags(0)

	// parse flags
	device := flag.String("d", "", "device")
	configPath := flag.String("e", "", "config path")
	flag.Parse()
	config := loadConfig(*configPath)

	// start packet capturing
	source, err := probe.NewProbeSource(*device, config.ApName)
	if err != nil {
		log.Fatal(err)
	}
	defer source.Close()

	for record := range source.Records() {
		log.Println(record)

		bytes, err := json.Marshal(record)
		if err != nil {
			continue
		}
		log.Println(string(bytes))
	}
}

//func postRecord(record *ProbeRecord) {
//
//    // set recover for net/http panic
//    defer func() {
//        if err := recover(); err != nil {
//            fmt.Println("recover: ", err)
//        }
//    }()
//
//    res, _ := http.PostForm(config.ApiUrl, record.Values())
//    defer res.Body.Close()
//
//    buf := new(bytes.Buffer)
//    buf.ReadFrom(res.Body)
//    fmt.Println(buf.String())
//}
