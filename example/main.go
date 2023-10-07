package main

import (
	"fmt"
	"log"
	"net"
	"os"

	"github.com/florianl/go-tc"
	"github.com/mdlayher/netlink"
	"golang.org/x/sys/unix"
)

func main() {
	// open a rtnetlink socket
	rtnl, err := tc.Open(&tc.Config{Logger: log.Default()})
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not open rtnetlink socket: %v\n", err)
		return
	}
	defer func() {
		if err := rtnl.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "could not close rtnetlink socket: %v\n", err)
		}
	}()

	// For enhanced error messages from the kernel, it is recommended to set
	// option `NETLINK_EXT_ACK`, which is supported since 4.12 kernel.
	//
	// If not supported, `unix.ENOPROTOOPT` is returned.

	err = rtnl.SetOption(netlink.ExtendedAcknowledge, true)
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not set option ExtendedAcknowledge: %v\n", err)
		return
	}

	// get all the qdiscs from all interfaces
	qdiscs, err := rtnl.Qdisc().Get()
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not get qdiscs: %v\n", err)
		return
	}

	ifName := "eth0"
	vfRep, err := net.InterfaceByName(ifName)
	if err != nil {
		log.Fatalf("failed to get interface %s: %v", ifName, err)
	}

	for _, qdisc := range qdiscs {
		if int(qdisc.Ifindex) != vfRep.Index {
			continue
		}
		iface, err := net.InterfaceByIndex(int(vfRep.Index))
		if err != nil {
			fmt.Fprintf(os.Stderr, "could not get interface from id %d: %v", qdisc.Ifindex, err)
			return
		}
		fmt.Printf("%20s\t%+v\n", iface.Name, qdisc)
	}
	fmt.Println()
	fmt.Println()
	fmt.Println()

	getFilterMsg := &tc.Msg{
		Family:  unix.AF_UNSPEC,
		Ifindex: uint32(vfRep.Index),
		Handle:  uint32(5),
		Parent:  tc.HandleRoot,
	}
	filters, err := rtnl.Filter().Get(getFilterMsg)
	if err != nil {
		log.Fatalf("failed to get filters: %v", err)
	}
	for _, filter := range filters {
		if filter.Flower != nil {
			//if filter.Flower.KeyIPv4Src != nil {
			//	fmt.Printf("flower ipv4 src is %s\n", *filter.Flower.KeyIPv4Src)
			//}
			//if filter.Flower.KeyIPv4Dst != nil {
			//	fmt.Printf("flower ipv4 dst is %s\n", *filter.Flower.KeyIPv4Dst)
			//}
			//if filter.Flower.keyIPv6Src != nil {
			//	fmt.Printf("flower ipv6 src is %s\n", *filter.Flower.keyIPv6Src)
			//}
			//if filter.Flower.keyIPv6Dst != nil {
			//	fmt.Printf("flower ipv6 dst is %s\n", *filter.Flower.keyIPv6Dst)
			//}
			fmt.Printf("flower key is %+v\n", *filter.Flower)
			for _, act := range *filter.Flower.Actions {
				fmt.Printf("action kind is %s\n", act.Kind)
				if act.Stats != nil {
					fmt.Printf("action %s stats is %d\n", act.Kind, act.Stats.Basic.Packets)
				}
			}
			fmt.Println()
		}
	}
}
