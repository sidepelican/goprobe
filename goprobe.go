package main

import (
	"log"
	"os"
	"io"
	"flag"
	"path"
	"encoding/json"
	"runtime"
	//"net/http"
	//"bytes"

	"github.com/BurntSushi/toml"
	"github.com/sidepelican/goprobe/probe"
	MQTT "github.com/eclipse/paho.mqtt.golang"
)

type Config struct {
	ApiUrl  string
	ApName  string
	MqttUri string
}

func exists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

func findDefaultConfigPath() string {
	if runtime.GOOS == "linux" {
		p := "/etc/goprobe/config.tml"
		if exists(p) {
			return p
		}
	}

	runPath, err := os.Executable()
	if err == nil {
		return path.Dir(runPath) + "/config.tml"
	}

	return ""
}

func loadConfig(path string) (config Config) {
	config.ApName = "undefined"

	if path == "" {
		path = findDefaultConfigPath()
		if path == "" {
			log.Println("cannot find config file path")
		} else {
			log.Println("load default config:", path)
		}
	}

	// decode const settings
	if _, err := toml.DecodeFile(path, &config); err != nil {
		log.Println(err)
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

	// setup MQTT (optional)
	var mqttClient MQTT.Client = nil
	if config.MqttUri != "" {
		opts := MQTT.NewClientOptions()
		opts.AddBroker(config.MqttUri)

		mqttClient = MQTT.NewClient(opts)
		if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
			log.Println("MQTT:", token.Error())
		} else {
			defer mqttClient.Disconnect(250)
			log.Println("MQTT Publisher Started")
		}
	}

	for record := range source.Records() {
		log.Println(record)

		bytes, err := json.Marshal(record)
		if err != nil {
			continue
		}

		if mqttClient != nil {
			t := mqttClient.Publish("/goprobe", 2, false, bytes)
			t.Wait()
		}
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
