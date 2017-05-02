package main

import (
    "log"
    //"net/http"
    //"bytes"
    //"github.com/BurntSushi/toml"
    "flag"

    "github.com/sidepelican/goprobe/probe"
)

type Config struct {
    ApiUrl string
    Apname string
}

var config Config


func main() {
    //// logging setup
    //f, err := os.OpenFile("/var/log/goprobe.log", os.O_APPEND | os.O_CREATE | os.O_RDWR, 0666)
    //if err != nil {
    //    println("error opening file: %v", err)
    //}
    //defer f.Close()
    //log.SetOutput(io.MultiWriter(f, os.Stdout)) // assign it to the standard logger
    log.SetFlags(0)

    // parse flags
    device := flag.String("d", "", "device")
    flag.Parse()

    // decode const settings
    //if _, err := toml.DecodeFile("config.tml", &config); err != nil {
    //    fmt.Printf("load config.tml failed: %s\n", err)
    //    return
    //}

	// start packet capturing
	source, err := probe.NewProbeSource(*device)
	if err != nil{
		log.Fatal(err)
	}
	defer source.Close()

    for record := range source.Records() {
        log.Println(record)
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
