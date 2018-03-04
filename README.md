# goprobe
probe request packet capturing utility in Golang

## Features
- Easy install as daemon to your device (mac, raspberry Pi)
- Find and set up monitor mode interface automatically
- Autorotate logging
- Send caputured probe request to MQTT or HTTP server

## Build and Install
for Raspberri Pi (Raspbian)
```
$ sudo apt-get install libpcap-dev
$ go get github.com/sidepelican/goprobe
$ cd ${GOPATH}/src/github.com/sidepelican/goprobe
$ go build
$ sudo ./goprobe install    # regist as a daemon
```

## Usage

- Run (not as a daemon)
```
$ sudo ./goprobe
```


### Config
copy and rename `config_sample.tml` to `config.tml`. And see inside it.