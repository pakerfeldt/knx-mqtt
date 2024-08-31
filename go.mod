module github.com/pakerfeldt/knx-mqtt

go 1.22.5

require (
	github.com/eclipse/paho.mqtt.golang v1.5.0
	github.com/rs/zerolog v1.33.0
	github.com/vapourismo/knx-go v0.0.0-20240623212929-3b325e3f5dcf
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/gorilla/websocket v1.5.3 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	golang.org/x/net v0.27.0 // indirect
	golang.org/x/sync v0.7.0 // indirect
	golang.org/x/sys v0.24.0 // indirect
	golang.org/x/text v0.16.0 // indirect
)

replace github.com/vapourismo/knx-go => github.com/pakerfeldt/knx-go v0.0.1
