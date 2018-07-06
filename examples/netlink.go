// For forwarding testing call
// "insmod ./x86_64-native-linuxapp-gcc/kmod/rte_kni.ko lo_mode=lo_mode_fifo_skb"
// from DPDK directory before compiling this test. It will make a loop of packets
// inside KNI device and receive from KNI will receive all packets that were sent to KNI.

// For ping testing call
// "insmod ./x86_64-native-linuxapp-gcc/kmod/rte_kni.ko"
// from DPDK directory before compiling this test. Use --ping option.

// Other variants of rte_kni.ko configuration can be found here:
// http://dpdk.org/doc/guides/sample_app_ug/kernel_nic_interface.html

// Need to call "ifconfig myKNI 111.111.11.11" while running this example to allow other applications
// to receive packets from "111.111.11.11" address

package main

import (
	"flag"
	"fmt"
	"net"
	"time"

	"github.com/intel-go/nff-go/common"
	"github.com/intel-go/nff-go/flow"
	"github.com/intel-go/nff-go/packet"
	"golang.org/x/sys/unix"

	"github.com/vishvananda/netlink"
)

var ping bool

// Map destination subnet to GW IP address
var routeCache map[string]net.IP

func main() {

	inport := flag.Uint("inport", 0, "port for receiver")
	outport := flag.Uint("outport", 0, "port for sender")
	kniport := flag.Uint("kniport", 0, "port for kni")
	flag.BoolVar(&ping, "ping", false, "use this for pushing only ARP and ICMP packets to KNI")
	flag.Parse()

	config := flow.Config{
		// Is required for KNI
		NeedKNI: true,
		CPUList: "0-7",
	}

	initRouteCache()

	go updateRouteCache()

	flow.CheckFatal(flow.SystemInit(&config))
	// port of device, name of device
	kni, err := flow.CreateKniDevice(uint16(*kniport), "myKNI")
	flow.CheckFatal(err)

	inputFlow, err := flow.SetReceiver(uint16(*inport))
	flow.CheckFatal(err)

	toKNIFlow, err := flow.SetSeparator(inputFlow, pingSeparator, nil)
	flow.CheckFatal(err)

	flow.CheckFatal(flow.SetSenderKNI(toKNIFlow, kni))
	fromKNIFlow := flow.SetReceiverKNI(kni)

	outputFlow, err := flow.SetMerger(inputFlow, fromKNIFlow)
	flow.CheckFatal(err)
	flow.CheckFatal(flow.SetSender(outputFlow, uint16(*outport)))

	flow.CheckFatal(flow.SystemStart())
}

func pingSeparator(current *packet.Packet, ctx flow.UserContext) bool {
	if ping == false {
		// All packets will go to KNI.
		// You should use lo_mode=lo_mode_fifo_skb for looping back these packets.
		return false
	}
	ipv4, ipv6, arp := current.ParseAllKnownL3()
	if arp != nil {
		return false
	} else if ipv4 != nil {
		if ipv4.NextProtoID == common.ICMPNumber {
			return false
		}
	} else if ipv6 != nil {
		if ipv6.Proto == common.ICMPNumber {
			return false
		}
	}
	return true
}

func initRouteCache() {
	common.LogDebug(common.Debug, "------- Init Route Cache -------")

	routes, err := netlink.RouteList(nil, netlink.FAMILY_ALL)
	if err != nil {
		common.LogFatal(common.Debug, "Cannot list routs")
	}

	routeCache = make(map[string]net.IP)
	for k, v := range routes {
		common.LogDebug(common.Debug, "Add route: ", k, v)
		dst := "0.0.0.0/0"
		if v.Dst != nil {
			dst = v.Dst.String()
		}
		routeCache[dst] = v.Gw
	}
	printRouteCache()
}

func printKernelRouteTable() {
	common.LogDebug(common.Debug, "------- Kernel Route Table -------")
	routes, err := netlink.RouteList(nil, netlink.FAMILY_ALL)
	if err != nil {
		common.LogFatal(common.Debug, "Cannot list routs")
	}
	for k, v := range routes {
		common.LogDebug(common.Debug, "Route: ", k, v.Dst, v.Gw)
		dst := "0.0.0.0/0"
		if v.Dst != nil {
			dst = v.Dst.String()
		}
		routeCache[dst] = v.Gw
	}
}

func printRouteCache() {
	common.LogDebug(common.Debug, "------- Route Cache -------")
	for k, v := range routeCache {
		fmt.Printf("key=%s, val=%v\n", k, v)
	}
}

func updateRouteCache() {
	ch := make(chan netlink.RouteUpdate)
	done := make(chan struct{})

	defer close(done)

	if err := netlink.RouteSubscribe(ch, done); err != nil {
		common.LogFatal(common.Debug, "Cannot subscribe:", err)
	}

	for {
		timeout := time.After(2 * time.Second)
		select {
		case update := <-ch:
			dst := update.Route.Dst
			gw := update.Route.Gw

			if update.Type == unix.RTM_NEWROUTE {
				common.LogDebug(common.Debug, "========== New route added! ", update)
				tmpdst := "0.0.0.0/0"
				if dst != nil {
					tmpdst = dst.String()
				}
				routeCache[tmpdst] = gw
			}

			if update.Type == unix.RTM_DELROUTE && dst != nil && gw != nil {
				common.LogDebug(common.Debug, "========== Route deleted ", update)
				if _, ok := routeCache[dst.String()]; ok {
					delete(routeCache, dst.String())
				}
			}
		case <-timeout:
			common.LogDebug(common.Debug, "===== No changes after 2s =====")
			printKernelRouteTable()
			printRouteCache()
		}
	}
}
