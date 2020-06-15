package main

import (
	"errors"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/sparrc/go-ping"
)

func main() {

	const art = `


 +-+-+-+-+-+-+-+-+-+-+-+
 |S|u|b|n|e|t|-|P|I|N|G|
 +-+-+-+-+-+-+-+-+-+-+-+-+-+-+
           |-| |L|a|k|s|h|y|a|
           +-+ +-+-+-+-+-+-+-+

        `

	const banner = `
        
        This Program Pings(ICMP) all the given IP's in the range concurrently
        with a timeout of 5 seconds (~ approx 5 Packets) and checks if the IP is reachable.
        
        `
	fmt.Println(art)
	fmt.Println(banner)

	c := make(chan string)

	if len(os.Args) != 3 {
		fmt.Println(errors.New("check your IP address and Mask"))
		return
	}
	ip := net.ParseIP(os.Args[1])
	netmask := net.ParseIP(os.Args[2])
	hostrange := int(255 - netmask[15])

	fmt.Printf("\n\nThere are %v IP's in the given Subnet.\n", hostrange)
	fmt.Printf("\nPinging all IP's in the given range......\n\n")
	fmt.Print("\n================================================================================\n")

	var hostslice [][]byte
	for i := 0; i <= hostrange; i++ {
		hostslice = append(hostslice, []byte{ip[12], ip[13], ip[14], byte(i)})
	}
	var hostslicestring []string
	for _, v := range hostslice[1:] {
		sip := net.IPv4(v[0], v[1], v[2], v[3])
		hostslicestring = append(hostslicestring, sip.String())
	}
	for _, v := range hostslicestring {
		go reachcheck(v, c)
	}

	for i := 0; i <= len(hostslicestring)-1; i++ {
		fmt.Println(<-c)
	}
}

func reachcheck(s string, c chan string) {
	pinger, err := ping.NewPinger(s)
	if err != nil {
		fmt.Println("Error in Address", s)
		return
	}
	pinger.SetPrivileged(true)
	pinger.Timeout = 5 * time.Second
	pinger.Run()
	if pinger.PacketsRecv == 0 {
		c <- "Unreachable or Not Provisioned ====>   " + s
		return
	} else {
		c <- s + " is REACHABLE"
		return
	}

}
