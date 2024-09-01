# KNX / MQTT bridge in Go

Welcome to the KNX / MQTT Bridge! This versatile application, crafted in Go, facilitates seamless integration between your KNX installation and MQTT. Perfect for smart home enthusiasts and IoT integrators, this bridge offers customizable and efficient connectivity.

## Features
* Customizable Integration: Connect your KNX system with MQTT effortlessly.
* Automatic Type Conversion: Provide an XML export from ETS, and the app will handle type conversions automatically.
* Flexible MQTT Topics: Publish KNX events to MQTT topics formatted as `knx/x/y/z`, where `x/y/z` represents the group address. You can also configure the `knx/` prefix and use names for group address sections.

## KNX XML Export
To get the most out of this application, provide an ETS XML export of your group addresses and their datapoint types. This enables automatic conversion between raw types and ensures precise data handling. Without this export, you’ll be limited to handling raw bytes only.

## Configuration
For a detailed guide on setting up and customizing the KNX/MQTT bridge, refer to the [example configuration file](https://github.com/pakerfeldt/knx-mqtt/blob/main/config.example.yaml). This file is thoroughly documented and provides comprehensive instructions for tailoring the bridge to your specific needs.

Please note that the default configuration logs at the info level, which does not include detailed logs of successfully read KNX messages. During the integration phase, you may want to adjust the log level to debug to see incoming messages and gain better visibility into the bridge's operations.

## Usage
You can build the application yourself, but using Docker is often more convenient. Here’s how you can get started with Docker:

### Docker
```
docker run --rm -it -v $(pwd)/config.yaml:/app/config.yaml -v $(pwd)/knx.xml:/app/knx.xml pakerfeldt/knx-mqtt:latest
```

### Docker Compose
```
services:
  knx-mqtt:
    image: pakerfeldt/knx-mqtt:latest
    restart: unless-stopped
    volumes:
      - ./config.yaml:/app/config.yaml
      - ./knx.xml:/app/knx.xml
```

## KNX to MQTT

### MQTT message format
The app supports several payload formats for MQTT messages:

| Type    | Description | Requires KNX XML |
| -------- | ------- | ------- |
| `value`  | String representation of the value (e.g., `24.52`) | Yes |
| `value-with-unit` | String representation including the unit (e.g., `24.42 °C`) | Yes |
| `json` | JSON representation of the message | Yes |
| `bytes` | Raw bytes as specified by KNX | No |

When sending JSON messages, you can include the following fields:

| Field    | Description |
| -------- | ------- |
| `bytes`  | Raw bytes encoded in Base64 |
| `name` | Name of the group address |
| `value` | String representation of the value |
| `unit` | Associated unit of the value |

## MQTT to KNX

KNX group addresses can be referred to using either their group address `knx/x/y/z/` or their full name, 
where x and y are the names of the group ranges and z is the name of the actual group address.

### Sending read requests
To send a read request, write to `knx/x/y/z/read` with any payload.

### Writing raw bytes to an address
To write to a group address with its raw bytes, send a message to `knx/x/y/z/write-bytes` with the bytes as payload.

### Writing value as a string to an address
To write to a group address using a string representation, send a message to `knx/x/y/z/write` with the value as a string.
E.g. `"25.35"`, `"true"`.

## Migrating from knx-mqtt-bridge
If you’re transitioning from [knx-mqtt-bridge](https://github.com/pakerfeldt/knx-mqtt-bridge), the NodeJS version of this bridge, you’ll find the migration relatively straightforward. The configurations are similar enough to make the transition smooth.

## Caveats
This app does not yet support MQTT over TLS.
