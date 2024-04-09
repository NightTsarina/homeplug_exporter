# HomePlug exporter for Prometheus

Prometheus exporter for [HomePlug](https://en.wikipedia.org/wiki/HomePlug)
Power Line Communication (PLC) devices.

It collects and exports metrics about the HomePlug network and stations.

Tested with TP-Link and Devolo devices but should work with any device using a
Qualcomm Atheros chipset such as QCA6410, QCA7000, and QCA7420.

Currently it does not support other chipsets as it uses Atheros-specific
messages to gather statistics.

**Note:** Due to the use of github.com/mdlayher/packet for low-level network
functionality, homeplug_exporter will only work on Linux systems. Attempting to
run it on non-Linux systems will result in "packet: not implemented on [GOOS]"
errors.

# Running

Command-line flags:

```
usage: homeplug_exporter [<flags>]

Flags:
  -h, --help                 Show context-sensitive help (also try --help-long
                             and --help-man).
      --telemetry.address=":9702"
                             Address on which to expose metrics.
      --telemetry.endpoint="/metrics"
                             Path under which to expose metrics.
      --interface=INTERFACE  Interface to search for Homeplug devices.
      --destaddr=local       Destination MAC address for Homeplug devices.
                             Accepts 'local', 'all', and 'broadcast' as aliases.
      --log.level=info       Only log messages with the given severity or above.
                             One of: [debug, info, warn, error]
      --log.format=logfmt    Output format of log messages. One of: [logfmt,
                             json]
      --version              Show application version.
```

The `destaddr` parameter specifies the MAC address of a HomePlug device or one
of these aliases:

 * `all`, `broadcast`  
    A synonym for the Ethernet broadcast address: `ff:ff:ff:ff:ff:ff`.
    All devices, whether local, remote, or foreign[^1] recognize messages sent to
    this address.

 * `local` (default)  
    A synonym for the Qualcomm Atheros vendor specific Local Management Address
    (LMA): `00:b0:52:00:00:01`.  All local Atheros devices recognize this
    address but remote and foreign devices do not.

Note that the default destination (`local`) will only find devices on the near
side of a PLC connection.

Once a reply is obtained from a device, the exporter directly queries all other
devices connected to the same HomePlug network.

[^1]: A "local device" is any device at the near end of a PLC connection.  
  A "remote device" is any device at the far end of a PLC connection.  
  A "foreign device" is any device not manufactured by Atheros.

## Permissions

The exporter needs to access raw sockets, so it needs to be run as root, or
with `cap_net_raw` capability. To avoid running it as root, set the
capabilities on the binary as follows:

```
sudo setcap cap_net_raw=eip /path/to/binary
```

## Using Docker

**NOTE:** The HomePlug protocol uses raw Ethernet frames, and must be run with `--net=host`
on the same layer 2 network segment as at least one HomePlug device.

```
docker build -t homeplug_exporter .
docker run --rm --detach --name=homeplug_exporter --net=host homeplug_exporter
```

# Metrics

```
# HELP homeplug_exporter_build_info A metric with a constant '1' value labeled by version, revision, branch, and goversion from which homeplug_exporter was built.
# TYPE homeplug_exporter_build_info gauge
homeplug_exporter_build_info{branch="main",goversion="go1.22.1",revision="",version="0.4.0"} 1

# HELP homeplug_network_info Logical network information
# TYPE homeplug_network_info gauge
homeplug_network_info{cco_addr="de:ad:be:ef:00:01",cco_tei="1",device_addr="de:ad:be:ef:00:01",nid="52:de:ad:be:ef:00:01",role="CCO",snid="6",tei="1"} 1
homeplug_network_info{cco_addr="de:ad:be:ef:00:01",cco_tei="1",device_addr="de:ad:be:ef:00:02",nid="52:de:ad:be:ef:00:01",role="STA",snid="6",tei="2"} 1
# HELP homeplug_station_rx_rate_bytes Average PHY Rx data rate
# TYPE homeplug_station_rx_rate_bytes gauge
homeplug_station_rx_rate_bytes{device_addr="de:ad:be:ef:00:01",nid="52:de:ad:be:ef:00:01",peer_addr="de:ad:be:ef:00:02"} 1.86e+07
homeplug_station_rx_rate_bytes{device_addr="de:ad:be:ef:00:02",nid="52:de:ad:be:ef:00:01",peer_addr="de:ad:be:ef:00:01"} 1.18e+06
# HELP homeplug_station_tx_rate_bytes Average PHY Tx data rate
# TYPE homeplug_station_tx_rate_bytes gauge
homeplug_station_tx_rate_bytes{device_addr="de:ad:be:ef:00:01",nid="52:de:ad:be:ef:00:01",peer_addr="de:ad:be:ef:00:02"} 1.18e+06
homeplug_station_tx_rate_bytes{device_addr="de:ad:be:ef:00:02",nid="52:de:ad:be:ef:00:01",peer_addr="de:ad:be:ef:00:01"} 1.86e+07
```
