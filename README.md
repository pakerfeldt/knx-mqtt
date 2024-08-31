# KNX / MQTT Bridge in Go

Welcome to the KNX / MQTT Bridge application! This powerful tool, written in Go, allows seamless integration between your KNX installation and MQTT. Whether you’re looking to enhance your smart home system or integrate with other IoT devices, this bridge is highly customizable to fit your needs.

## Features

- **Customizable Integration**: Easily bridge your KNX setup with MQTT.
- **Automatic Type Conversion**: Simply provide an XML export of your group addresses from ETS, and this app will automatically convert them to the appropriate types.
- **Flexible Topic Mapping**: KNX events can be published to MQTT topics in the format `knx/x/y/z`, where `x/y/z` represents the group address. The `knx/` prefix is configurable. You can also publish events to topics named after your group address sections.

## KNX XML Export

To make the most of this application, it is highly recommended to provide an ETS XML export of all your group addresses and their corresponding datapoint types. This allows automatic conversion of values between their raw types and ensures accurate data handling. Without this export, you will be limited to sending and receiving raw bytes.

## Configuration

The example configuration file is thoroughly documented to help you get started. Please refer to it for detailed setup instructions and options.

## Usage

You can build the application yourself, but using Docker is often more convenient. Here’s how you can get started with Docker:

### Docker

Run the following command to start the container, mapping your configuration and XML files:

```bash
docker run --rm -it -v $(pwd)/config.yaml:/app/config.yaml -v $(pwd)/knx.xml:/app/knx.xml pakerfeldt/knx-mqtt:latest
```

### Docker Compose
If you prefer Docker Compose, use the following configuration in your docker-compose.yml file:
```
services:
  knx-mqtt:
    image: pakerfeldt/knx-mqtt:latest
    restart: unless-stopped
    volumes:
      - /srv/knx-mqtt/config.yaml:/app/config.yaml
      - /srv/knx-mqtt/knx.xml:/app/knx.xml
```

### Getting Help
If you encounter any issues or have questions, feel free to open an issue on GitHub or reach out for support. Happy bridging!