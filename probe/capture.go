package probe

import (
	"fmt"
	"log"
	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
	"github.com/google/gopacket/layers"
)

type ProbeSource struct {
	c    chan ProbeRecord
	stop chan bool
}

func (s *ProbeSource) Records() chan ProbeRecord {
	return s.c
}

func (s *ProbeSource) Close() {
	s.stop <- true
}

func NewProbeSource(device string, accessPointName string) (*ProbeSource, error) {

	var handle *pcap.Handle
	var err error
	if device == "" {
		handle, err = openAvailableMonitorModeInterface()
	} else {
		handle, err = openAsMonitorMode(device)
	}
	if err != nil {
		return nil, err
	}
	log.Printf("pcap version: %s\n", pcap.Version())

	source := &ProbeSource{
		c:    make(chan ProbeRecord),
		stop: make(chan bool),
	}

	go func() {
		defer handle.Close()
		packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
		for {
			select {
			case packet := <-packetSource.Packets():
				// decode and find ProbeRequest
				probeLayer := packet.Layer(layers.LayerTypeDot11MgmtProbeReq)
				if probeLayer == nil {
					continue
				}

				dot11 := packet.Layer(layers.LayerTypeDot11).(*layers.Dot11)
				radioTap := packet.Layer(layers.LayerTypeRadioTap).(*layers.RadioTap)

				source.c <- ProbeRecord{
					Timestamp:  packet.Metadata().Timestamp.Unix(),
					Mac:        dot11.Address2.String(),
					Rssi:       int(radioTap.DBMAntennaSignal),
					SequenceId: int(dot11.SequenceNumber),
					ApName:     accessPointName,
				}
			case <-source.stop:
				break
			}
		}
	}()

	return source, nil
}

func openAvailableMonitorModeInterface() (*pcap.Handle, error) {

	// find useable devices
	ifs, err := pcap.FindAllDevs()
	if len(ifs) == 0 {
		return nil, fmt.Errorf("no devices found: %s", err)
	}

	// try any interfaces to monitor
	errs := ""
	for i, iface := range ifs {
		handle, err := openAsMonitorMode(iface.Name)
		if err != nil {
			errs += fmt.Sprintf("dev %d: %s (%s)\n%v\n", i+1, iface.Name, iface.Description, err)
			continue
		}
		log.Println("used interface:", iface.Name)
		return handle, nil
	}

	return nil, fmt.Errorf("failed to find monitor mode available interface\n%s", errs)
}

func openAsMonitorMode(device string) (*pcap.Handle, error) {

	inactive, err := pcap.NewInactiveHandle(device)
	if err != nil {
		return nil, fmt.Errorf("NewInactiveHandle(%s) failed: %s", device, err)
	}
	defer inactive.CleanUp()

	// change mode to monitor
	if err := inactive.SetRFMon(true); err != nil {
		return nil, fmt.Errorf("SetRFMon failed: %s", err)
	}

	// create the actual handle by calling Activate:
	handle, err := inactive.Activate() // after this, inactive is no longer valid
	if handle == nil {
		return nil, fmt.Errorf("Activate(%s) failed: %s", device, err)
	}

	return handle, nil
}
