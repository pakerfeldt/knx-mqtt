# Version 1.6
- New minimal docker image introduced
- App will exit out if connection to MQTT is lost to allow for docker to automatically restart

# Version 1.5
- Add option to include address and dpt in MQTT message
- Allow for specifying which interface to use when using multicast 
- BUGFIX: Do not emit Read commands as value over MQTT
- Add option to include/exclude read commands from KNX to MQTT

# Version 1.4
- Support MQTT over TLS.

# Version 1.3
- Gracefully tries to reconnect to KNX and MQTT if connection is lost.
- Fixed a bug where it would not send to KNX if the address existed in the ETS export and outgoing message type was set to 'bytes'.
- This release is a complete restructuring of the code base.

# Version 1.2
- Replaces / with _ in GroupAddress Names to avoid unwanted MQTT topic separations.

# Version 1.1
- Added configuration parameter emitValueAsString. When set to `false`, values will preserve its type when sent as `json`. If set to `true`, behaviour is compatible with previous version.
