package main

import (
    "fmt"
    "log"
    //"net/http"
    //"bytes"
    //"github.com/BurntSushi/toml"
    "flag"

    "github.com/google/gopacket"
    "github.com/google/gopacket/pcap"
    "github.com/google/gopacket/layers"
    "github.com/sidepelican/goprobe/probe"
)

type Config struct {
    ApiUrl string
    Apname string
}

var config Config

func openAvailableMonitorModeInterface() (*pcap.Handle, error) {

    // find useable devices
    ifs, err := pcap.FindAllDevs()
    if len(ifs) == 0 {
        return nil, fmt.Errorf("no devices found : %s\n", err)
    }

    // try any interfaces to monitor
    for _, iface := range ifs {
        handle, err := openAsMonitorMode(iface.Name)
        if err != nil {
            continue
        }
        fmt.Println("open interface", iface.Name)
        return handle, nil
    }

    errs := ""
    for i, iface := range ifs {
        errs += fmt.Sprintf("\ndev %d: %s (%s)", i+1, iface.Name, iface.Description)
    }

    return nil, fmt.Errorf("failed to find monitor mode available interface.%s", errs)
}

func openAsMonitorMode(device string) (*pcap.Handle, error) {

    inactive, err := pcap.NewInactiveHandle(device)
    if err != nil {
        return nil, fmt.Errorf("NewInactiveHandle(%s) failed: %s\n", device, err)
    }
    defer inactive.CleanUp()

    // change mode to monitor
    if err := inactive.SetRFMon(true); err != nil {
        return nil, fmt.Errorf("SetRFMon failed: %s\n", err)
    }

    // create the actual handle by calling Activate:
    handle, err := inactive.Activate() // after this, inactive is no longer valid
    if handle == nil {
        return nil, fmt.Errorf("Activate(%s) failed: %s\n", device, err)
    }

    return handle, nil
}

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

    // device name must be set
    var handle *pcap.Handle
    var err error
    if *device == "" {
        handle, err = openAvailableMonitorModeInterface()
    } else {
        handle, err = openAsMonitorMode(*device)
    }
    if err != nil {
        log.Fatal(err)
    }
    defer handle.Close()
    fmt.Printf("pcap version: %s\n", pcap.Version())

    // decode const settings
    //if _, err := toml.DecodeFile("config.tml", &config); err != nil {
    //    fmt.Printf("load config.tml failed: %s\n", err)
    //    return
    //}

    // start packet capturing
    packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
    for packet := range packetSource.Packets() {

        // decode and find ProbeRequest
        probeLayer := packet.Layer(layers.LayerTypeDot11MgmtProbeReq)
        if probeLayer == nil {
            //fmt.Println("other packet found", packet)
            continue
        }

        dot11 := packet.Layer(layers.LayerTypeDot11).(*layers.Dot11)
        radioTap := packet.Layer(layers.LayerTypeRadioTap).(*layers.RadioTap)

        // send a record to DB
        r := probe.ProbeRecord{
            Time:       packet.Metadata().Timestamp,
            Mac:        dot11.Address2.String(),
            Rssi:       int(radioTap.DBMAntennaSignal),
            SequenceId: int(dot11.SequenceNumber),
            ApName:     config.Apname,
        }
        //postRecord(r)

        log.Println(r)
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
