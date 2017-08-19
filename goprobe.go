package main

import (
	"log"
	"fmt"
	"os"
	"io"
	"path"
	"encoding/json"
	"runtime"
	//"net/http"
	//"bytes"

	"github.com/BurntSushi/toml"
	MQTT "github.com/eclipse/paho.mqtt.golang"

	"github.com/sidepelican/goprobe/probe"
)

const topic = "/goprobe"

type Config struct {
	Device  string
	ApiUrl  string
	ApName  string
	MqttUri string
}

func mainLoop() {
	defer func() {
		if err := recover(); err != nil {
			log.Println("recover: ", err)
		}
	}()

	// logging setup
	log.SetFlags(0)
	const logPath = "/var/log/goprobe.log"
	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		log.Println("error opening file: %v", err)
	} else {
		log.Println("logging to ", logPath)
		defer f.Close()
		log.SetOutput(io.MultiWriter(f, os.Stdout)) // assign it to the standard logger
	}

	config := loadConfig()

	// start packet capturing
	source, err := probe.NewProbeSource(config.Device)
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

		// set sensor environment information
		if config.ApName != "" {
			record.ApName = config.ApName
		}

		log.Println(record.String())

		// send data as json
		bytes, err := json.Marshal(record)
		if err != nil {
			continue
		}

		if mqttClient != nil {
			go publishMqtt(mqttClient, bytes)
		}
	}
}

func publishMqtt(client MQTT.Client, bytes []byte) {
	defer func() {
		if err := recover(); err != nil {
			log.Println("recover: ", err)
		}
	}()

	t := client.Publish(topic, 2, false, bytes)
	t.Wait()
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

func exists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

func findConfigPath() (string, error) {

	const configFileNAme = "config.tml"
	errret := fmt.Errorf("%s not found at: ", configFileNAme)

	// static path
	if runtime.GOOS == "linux" {
		p := "/etc/goprobe/" + configFileNAme
		if exists(p) {
			return p, nil
		}
		errret = fmt.Errorf("%v\n\t%v", errret, p)
	}

	// runpath
	runPath, err := os.Executable()
	if err == nil {
		p := path.Dir(runPath) + "/" + configFileNAme
		if exists(p) {
			return p, nil
		}
		errret = fmt.Errorf("%v\n\t%v", errret, p)
	}

	// current dir
	pwd, err := os.Getwd()
	if err == nil {
		p := pwd + "/" + configFileNAme
		if exists(p) {
			return p, nil
		}
		errret = fmt.Errorf("%v\n\t%v", errret, p)
	}

	return "", errret
}

func loadConfig() (config Config) {

	path, err := findConfigPath()
	if err != nil {
		log.Println(err)
		return
	}
	log.Println("load config:", path)

	// decode const settings
	if _, err := toml.DecodeFile(path, &config); err != nil {
		log.Println(err)
		return
	}
	return
}
