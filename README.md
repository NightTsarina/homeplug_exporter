# HomePlug Power Line Communication Exporter for Prometheus

Exports HomePlug PLC network and station statistics via HTTP for Prometheus consumption.

Help on flags:

```
usage: homeplug_exporter [<flags>]

Flags:
  -h, --help                 Show context-sensitive help (also try --help-long and --help-man).
      --telemetry.address=":9702"
                             Address on which to expose metrics.
      --telemetry.endpoint="/metrics"
                             Path under which to expose metrics.
      --interface=INTERFACE  Interface to search for Homeplug devices.
      --destaddr=local       Destination MAC address for Homeplug devices. Accepts 'local', 'all', and 'broadcast' as aliases.
      --log.level=info       Only log messages with the given severity or above. One of: [debug, info, warn, error]
      --log.format=logfmt    Output format of log messages. One of: [logfmt, json]
      --version              Show application version.
```

Tested with TP-Link TL-PA4010, but should work with any device that supports HomePlug AV or better.

The default destination MAC address (`local`) will elicit a response from any
HomePlug devices on the local Layer 2 network segment.  This will NOT find
devices on the far side of a Power Line bridge. If you know the MAC address of
a device, including a device on the far side of a Power Line bridge, you may
override the destination address.

The `destaddr` parameter accepts these special values:

`all`
    Same as `broadcast`.

`broadcast`
    A synonym for the Ethernet broadcast address, `FF:FF:FF:FF:FF:FF`.
    All devices, whether local, remote or foreign recognize messages sent to this address.

`local`
    A synonym for the Qualcomm Atheros vendor specific Local Management Address (LMA), `00:B0:52:00:00:01`.
    All local Atheros devices recognize this address but remote and foreign devices do not.

A remote device is any device at the far end of a powerline connection.
A foreign device is any device not manufactured by Atheros.

# Running

## Using Docker

**NOTE:** The HomePlug protocol uses raw ethernet frames, and must be run with `--net=host`
on the same Layer 2 network segment as at least one HomePlug device.

```
docker run --rm --detach --name=homeplug_exporter --net=host brandond/homeplug_exporter
```

# Details

## Collectors

```
# HELP homeplug_exporter_build_info A metric with a constant '1' value labeled by version, revision, branch, and goversion from which homeplug_exporter was built.
# TYPE homeplug_exporter_build_info gauge
# HELP homeplug_network_id Logical network information
# TYPE homeplug_network_id gauge
# HELP homeplug_station_rx_rate_bytes Average PHY Rx data rate
# TYPE homeplug_station_rx_rate_bytes gauge
# HELP homeplug_station_tx_rate_bytes Average PHY Tx data rate
# TYPE homeplug_station_tx_rate_bytes gauge
```
