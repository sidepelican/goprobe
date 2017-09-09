package main

import (
	"log"
	"fmt"
	"os"
	"io"
	"path"
	"time"
	"encoding/json"
	"runtime"
	"net/http"
	"bytes"
	"io/ioutil"

	"github.com/sidepelican/goprobe/probe"
	"github.com/BurntSushi/toml"
	"github.com/lestrrat/go-file-rotatelogs"
	MQTT "github.com/eclipse/paho.mqtt.golang"
)

const (
	topic = "/goprobe"
	logPath = "/var/log/goprobe.log"
)

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
	rl, err := makeRotatelogs()
	if err != nil {
		log.Println("error setup rotatelogs: %v", err)
	} else {
		log.Println("logging to ", logPath)
		defer rl.Close()
		log.SetOutput(io.MultiWriter(rl, os.Stdout)) // assign it to the standard logger
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

		if config.ApiUrl != "" {
			go postRecord(config.ApiUrl, bytes)
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

func postRecord(url string, data []byte) {

    // set recover for net/http panic
    defer func() {
        if err := recover(); err != nil {
            fmt.Println("recover: ", err)
        }
    }()

	req, err := http.NewRequest(
		http.MethodPost,
		url,
		bytes.NewBuffer(data),
	)
	if err != nil {
		log.Println(err)
		return
	}

	// Content-Type setting
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Println(err)
		return
	}
	log.Println(string(body))
}

func exists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

func makeRotatelogs() (*rotatelogs.RotateLogs, error) {
	return rotatelogs.New(
		logPath + ".%Y%m%d",
		rotatelogs.WithLinkName(logPath),
		rotatelogs.WithRotationTime(time.Hour),
		rotatelogs.WithMaxAge(0),
	)
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

	// use hostname for Apname when it is empty
	if config.ApName == "" {
		config.ApName, _ = os.Hostname()
	}

	return
}
