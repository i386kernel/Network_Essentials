package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/sparrc/go-ping"
)

const art = `
 +-+-+-+-+-+-+-+-+-+-+-+-+-+
 |I|P|-|R|a|n|g|e|-|P|i|n|g|
 +-+-+-+-+-+-+-+-+-+-+-+-+-+
             |L|a|k|s|h|y|a|
             +-+-+-+-+-+-+-+
	  
        `

const banner = `
        
        This Program Pings(ICMP) all the given IP's in the range concurrently
        with a timeout of 5 seconds (~ approx 5 Packets) and checks if the IP is reachable.
        
        `

func pingIP(w io.Writer, start, end string) {

	c := make(chan string)
	startIPSplit := strings.Split(start, ".")
	endIPSplit := strings.Split(end, ".")

	startIPLastOctet, err := strconv.Atoi(startIPSplit[3])
	if err != nil {
		fmt.Println(err)
	}
	endIPLastOctet, err := strconv.Atoi(endIPSplit[3])
	if err != nil {
		fmt.Println(err)
	}
	var hostslice []string
	for i := startIPLastOctet; i <= endIPLastOctet; i++ {
		hostslice = append(hostslice, startIPSplit[0]+"."+startIPSplit[1]+"."+startIPSplit[2]+"."+strconv.Itoa(i))
	}
	hostslice = append(hostslice, "www.google.com")
	for _, v := range hostslice {
		go reachcheck(v, c)
	}
	for i := 0; i <= len(hostslice)-1; i++ {
		_, err := fmt.Fprintln(w, <-c)
		if err != nil {
			fmt.Println(err)
		}
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

func checkIP(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	startIP := ps.ByName("start")
	endIP := ps.ByName("end")
	fmt.Fprint(w, art)
	fmt.Fprint(w, banner)
	fmt.Fprintf(w, "\nStarting IP - %s\nEnding IP - %s\n", startIP, endIP)
	pingIP(w, startIP, endIP)
}

func main() {
	router := httprouter.New()
	router.GET("/:start/:end", checkIP)
	log.Fatal(http.ListenAndServe(":8000", router))
}
