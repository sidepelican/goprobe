package main

import (
	"log"
	"fmt"
	"os"
	"io"
	"flag"
	"path"
	"encoding/json"
	"time"
	"runtime"
	//"net/http"
	//"bytes"

	"github.com/BurntSushi/toml"
	"github.com/takama/daemon"
	MQTT "github.com/eclipse/paho.mqtt.golang"
	rotatelogs "github.com/lestrrat/go-file-rotatelogs"

	"github.com/sidepelican/goprobe/probe"
)

const (
	// name of the service
	name         = "goprobe"
	description  = "caputre probe requests"
	dependencies = "network.target"

	topic = "/goprobe"
)

type Config struct {
	ApiUrl  string
	ApName  string
	MqttUri string
}

type Service struct {
	daemon.Daemon
}

func (service *Service) Manage() (string, error) {

	// if received any kind of command, do it
	if len(os.Args) > 1 {
		command := os.Args[1]
		switch command {
		case "install":
			return service.Install()
		case "remove":
			return service.Remove()
		case "start":
			return service.Start()
		case "stop":
			return service.Stop()
		case "status":
			return service.Status()
		default:
			return fmt.Sprintf("Usage: %v [install | remove | start | stop | status]", os.Args[0]), nil
		}
	}

	// logging setup
	log.SetFlags(0)
	if runtime.GOOS == "linux" {
		logf, err := rotatelogs.New(
			"/var/log/goprobe.log.%Y%m%d%H%M",
			rotatelogs.WithLinkName("/var/log/goprobe.log"),
			rotatelogs.WithRotationTime(7*24*time.Hour),
		)
		if err != nil {
			log.Printf("failed to create rotatelogs: %s", err)
		} else {
			log.SetOutput(io.MultiWriter(logf, os.Stdout)) // assign it to the standard logger
		}
	}

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
			go publishMqtt(mqttClient, bytes)
		}
	}

	// never happen, but need to complete code
	return "", nil
}

func main() {
	srv, err := daemon.New(name, description, dependencies)
	if err != nil {
		log.Println("Error:", err)
		os.Exit(1)
	}
	service := &Service{srv}
	status, err := service.Manage()
	if err != nil {
		log.Println(status, "\nError:", err)
		os.Exit(1)
	}
	fmt.Println(status)
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

func findDefaultConfigPath() (string, error) {

	errret := fmt.Errorf("config.tml not found at: ")

	if runtime.GOOS == "linux" {
		p := "/etc/goprobe/config.tml"
		if exists(p) {
			return p, nil
		}
		errret = fmt.Errorf("%v\n\t%v", errret, p)
	}

	runPath, err := os.Executable()
	if err == nil {
		p := path.Dir(runPath) + "/config.tml"
		if exists(p) {
			return p, nil
		}
		errret = fmt.Errorf("%v\n\t%v", errret, p)
	}

	return "", errret
}

func loadConfig(path string) (config Config) {
	config.ApName = "undefined"

	if path == "" {
		var err error
		path, err = findDefaultConfigPath()
		if err != nil {
			log.Println(err)
			return
		}
		log.Println("load default path:", path)
	}

	// decode const settings
	if _, err := toml.DecodeFile(path, &config); err != nil {
		log.Println(err)
	}
	return
}
