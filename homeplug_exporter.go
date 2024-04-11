package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/alecthomas/kingpin/v2"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/mdlayher/packet"
	"github.com/prometheus/client_golang/prometheus"
	versioncollector "github.com/prometheus/client_golang/prometheus/collectors/version"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/promlog"
	"github.com/prometheus/common/promlog/flag"
	"github.com/prometheus/common/version"
	"github.com/prometheus/exporter-toolkit/web"
)

const (
	namespace = "homeplug"
)

var (
	stRole = [...]string{"STA", "PCO", "CCO"}

	listeningAddress = kingpin.Flag("telemetry.address", "Address on which to expose metrics.").Default(":9702").String()
	metricsEndpoint  = kingpin.Flag("telemetry.endpoint", "Path under which to expose metrics.").Default("/metrics").String()
	interfaceName    = kingpin.Flag("interface", "Interface to search for Homeplug devices.").String()
	destAddress      = MacAddress(kingpin.Flag("destaddr", "Destination MAC address for Homeplug devices. Accepts 'local', 'all', and 'broadcast' as aliases.").
				Default("local").HintOptions("local", "all", "broadcast"))

	logger log.Logger
)

type Exporter struct {
	iface *net.Interface
	conn  *packet.Conn
	dest  net.HardwareAddr
	mutex sync.Mutex

	txRate  *prometheus.Desc
	rxRate  *prometheus.Desc
	network *prometheus.Desc
}

func NewExporter(iface *net.Interface, conn *packet.Conn, dest net.HardwareAddr) *Exporter {
	return &Exporter{
		iface: iface,
		conn:  conn,
		dest:  dest,
		txRate: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "station", "tx_rate_bytes"),
			"Average PHY Tx data rate",
			[]string{"device_addr", "nid", "peer_addr"},
			nil),
		rxRate: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "station", "rx_rate_bytes"),
			"Average PHY Rx data rate",
			[]string{"device_addr", "nid", "peer_addr"},
			nil),
		network: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "network", "info"),
			"Logical network information",
			[]string{"device_addr", "nid", "snid", "tei", "role", "cco_addr", "cco_tei"},
			nil),
	}
}

func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- e.txRate
	ch <- e.rxRate
	ch <- e.network
}

func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	err := e.collect(ch)
	if err != nil {
		level.Error(logger).Log("msg", "Error scraping Homeplug", "err", err)
	}
}

func (e *Exporter) collect(ch chan<- prometheus.Metric) error {
	netinfos, err := get_homeplug_netinfo(e.iface, e.conn, e.dest)
	if err != nil {
		return err
	}

	for _, info := range netinfos {
		if len(info.Networks) == 0 {
			continue
		}
		for _, network := range info.Networks {
			ch <- prometheus.MustNewConstMetric(e.network, prometheus.GaugeValue, 1,
				info.Address.String(),
				network.NetworkID.String(),
				strconv.FormatUint(uint64(network.ShortID), 10),
				strconv.FormatUint(uint64(network.TEI), 10),
				stRole[network.Role],
				network.CCoAddress.String(),
				strconv.FormatUint(uint64(network.CCoTEI), 10))
		}

		network0 := info.Networks[0]
		for _, station := range info.Stations {
			ch <- prometheus.MustNewConstMetric(e.txRate, prometheus.GaugeValue,
				float64(uint64(station.TxRate)*1024*1024/8), info.Address.String(), network0.NetworkID.String(), station.Address.String())
			ch <- prometheus.MustNewConstMetric(e.rxRate, prometheus.GaugeValue,
				float64(uint64(station.RxRate)*1024*1024/8), info.Address.String(), network0.NetworkID.String(), station.Address.String())
		}
	}
	return nil
}

type HomeplugNetworkInfo struct {
	Address  net.HardwareAddr
	Networks []HomeplugNetworkStatus
	Stations []HomeplugStationStatus
}

func (n *HomeplugNetworkInfo) UnmarshalBinary(b []byte) error {
	if len(b) == 0 {
		return io.ErrUnexpectedEOF
	}

	// Number of AV logical networks.
	numNetworks := int(b[0])

	// Start offset of network / station data.
	o := 1

	for i := 0; i < numNetworks; i++ {
		var ns HomeplugNetworkStatus
		if err := (&ns).UnmarshalBinary(b[o:]); err != nil {
			return err
		}
		n.Networks = append(n.Networks, ns)
		o += 18 // size of HomeplugNetworkStatus struct

		for i := 0; i < int(ns.NumStations); i++ {
			var ss HomeplugStationStatus
			if err := (&ss).UnmarshalBinary(b[o:]); err != nil {
				return err
			}
			n.Stations = append(n.Stations, ss)
			o += 15 // size of HomeplugStationStatus struct
		}
	}

	return nil
}

// HomeplugNetworkStatus is analogous to the network struct defined in the open-plc-utils
// reference implementation (plc/NetInfo1.c).
type HomeplugNetworkStatus struct {
	NetworkID   networkID
	ShortID     uint8
	TEI         uint8
	Role        uint8
	CCoAddress  macAddr
	CCoTEI      uint8
	NumStations uint8
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface.
func (s *HomeplugNetworkStatus) UnmarshalBinary(b []byte) error {
	return binary.Read(bytes.NewReader(b), binary.LittleEndian, s)
}

type HomeplugStationStatus struct {
	Address        macAddr
	TEI            uint8
	BridgedAddress macAddr
	TxRate         uint8
	RxRate         uint8
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface.
func (s *HomeplugStationStatus) UnmarshalBinary(b []byte) error {
	return binary.Read(bytes.NewReader(b), binary.LittleEndian, s)
}

func main() {
	promlogConfig := &promlog.Config{}

	flag.AddFlags(kingpin.CommandLine, promlogConfig)
	kingpin.Version(version.Print("homeplug_exporter"))
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	logger = promlog.New(promlogConfig)

	level.Info(logger).Log("msg", "Starting homeplug_exporter", "version", version.Info())
	level.Info(logger).Log("msg", "Build context", "build_context", version.BuildContext())

	iface, err := getListenerInterface(*interfaceName)
	if err != nil {
		level.Error(logger).Log("msg", "Failed to get interface", "err", err)
		os.Exit(1)
	}

	conn, err := packet.Listen(iface, packet.Raw, etherType, nil)
	if err != nil {
		level.Error(logger).Log("msg", "Failed to listen", "err", err)
		os.Exit(1)
	}

	exporter := NewExporter(iface, conn, *destAddress)
	prometheus.MustRegister(exporter)
	prometheus.MustRegister(versioncollector.NewCollector("homeplug_exporter"))

	level.Info(logger).Log("msg", "Collector parameters", "destaddr", destAddress, "interface", iface.Name)
	level.Info(logger).Log("msg", "Starting HTTP server", "telemetry.address", *listeningAddress, "telemetry.endpoint", *metricsEndpoint)

	http.Handle(*metricsEndpoint, promhttp.Handler())
	if *metricsEndpoint != "/" {
		landingConfig := web.LandingConfig{
			Name:        "Homeplug Exporter",
			Description: "Prometheus Homeplug Exporter",
			Version:     version.Info(),
			Links: []web.LandingLinks{
				{
					Address: *metricsEndpoint,
					Text:    "Metrics",
				},
			},
		}
		landingPage, err := web.NewLandingPage(landingConfig)
		if err != nil {
			level.Error(logger).Log("err", err)
			os.Exit(1)
		}
		http.Handle("/", landingPage)
	}

	if err := http.ListenAndServe(*listeningAddress, nil); err != nil {
		level.Error(logger).Log("msg", "Failed to bind HTTP server", "err", err)
		os.Exit(1)
	}
}

func get_homeplug_netinfo(iface *net.Interface, conn *packet.Conn, dest net.HardwareAddr) ([]HomeplugNetworkInfo, error) {
	seen := make(map[string]bool, 0)
	ni := make([]HomeplugNetworkInfo, 0)
	ch := make(chan HomeplugNetworkInfo, 1)
	go read_homeplug(conn, ch)

	if err := writeQualcommReq(iface, conn, dest, nwInfoReq); err != nil {
		return nil, fmt.Errorf("writeQualcommReq failed: %w", err)
	}

ChanLoop:
	for {
		select {
		case n := <-ch:
			addr := n.Address.String()
			if seen[addr] {
				continue
			}
			ni = append(ni, n)
			seen[addr] = true
			// Query each remote station directly.
			for _, station := range n.Stations {
				if err := writeQualcommReq(iface, conn, net.HardwareAddr(station.Address[:]), nwInfoReq); err != nil {
					return nil, fmt.Errorf("writeQualcommReq failed: %w", err)
				}
			}

		case <-time.After(time.Second):
			break ChanLoop
		}
	}

	if len(ni) == 0 {
		return ni, nil
	}

	return ni, nil
}

func read_homeplug(conn *packet.Conn, ch chan<- HomeplugNetworkInfo) {
	for {
		addr, f, err := readFrame(conn)
		if err != nil {
			if !os.IsTimeout(err) {
				level.Error(logger).Log("msg", "failed to receive message", "err", err)
			}
			break
		}

		var h QualcommHdr
		if err := (&h).UnmarshalBinary(f.Payload[:qualcommHdrLen]); err != nil {
			level.Error(logger).Log("msg", "Failed to unmarshal homeplug frame", "err", err)
			continue
		}
		level.Debug(logger).Log(
			"msg", "Received homeplug frame",
			"from", addr,
			"version", fmt.Sprintf("%#x", h.Version),
			"mm_type", fmt.Sprintf("%#x", h.MMType),
			"vendor", fmt.Sprintf("%#x", h.Vendor),
			"payload", fmt.Sprintf("[% x]", f.Payload[qualcommHdrLen:]),
		)

		if h.MMType != nwInfoReq|mmTypeCnf {
			level.Error(logger).Log("msg", "Got unhandled MM type", "mm_type", fmt.Sprintf("%#x", h.MMType))
			continue
		}

		hni := HomeplugNetworkInfo{Address: f.Source}
		if err := (&hni).UnmarshalBinary(f.Payload[qualcommHdrLen:]); err != nil {
			level.Error(logger).Log("msg", "Failed to unmarshal network info frame", "err", err)
			continue
		}
		if len(hni.Networks) == 0 {
			level.Error(logger).Log("msg", "Ignoring isolated device", "device_addr", hni.Address)
			continue
		}

		for _, network := range hni.Networks {
			level.Debug(logger).Log(
				"msg", "Network found",
				"device_addr", hni.Address,
				"nid", network.NetworkID,
				"snid", network.ShortID,
				"tei", network.TEI,
				"role", stRole[network.Role],
				"cco_addr", network.CCoAddress,
				"cco_tei", network.CCoTEI,
			)
		}
		for _, station := range hni.Stations {
			level.Debug(logger).Log(
				"msg", "Connected station found",
				"device_addr", hni.Address,
				"peer_addr", station.Address,
				"bda", station.BridgedAddress,
				"tx_rate", station.TxRate,
				"rx_rate", station.RxRate,
			)
		}

		ch <- hni
	}
}

// getListenerInterface resolves a network interface name to a net.Interface. If the supplied
// ifName is empty, the first non-loopback interface which is up will be returned.
func getListenerInterface(ifName string) (*net.Interface, error) {
	if ifName == "" {
		ifaces, err := net.Interfaces()
		if err != nil {
			return nil, err
		}

		for _, iface := range ifaces {
			if iface.Flags&net.FlagLoopback == 0 && iface.Flags&net.FlagUp != 0 {
				return &iface, nil
			}
		}

		// No suitable interface found; return error.
		return nil, &net.OpError{Op: "route", Net: "ip+net", Err: errors.New("no such network interface")}
	}

	return net.InterfaceByName(ifName)
}
