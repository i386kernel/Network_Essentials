package main

import (
	"bytes"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"github.com/sparrc/go-ping"
	"html/template"
	"log"
	"net"
	"net/http"
	"path"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type PingStat struct {
	Available []string
	Occupied  []string
}

var stat = PingStat{}

func reachcheck(s string, wg *sync.WaitGroup) {

	// reset PingStat Struct to nil;
	stat.Available = nil
	stat.Occupied = nil

	defer wg.Done()
	pinger, err := ping.NewPinger(s)
	if err != nil {
		fmt.Println("Error in Address", s)
		return
	}
	pinger.SetPrivileged(true)
	pinger.Timeout = 5 * time.Second
	pinger.Run()
	if pinger.PacketsRecv == 0 {
		stat.Available = append(stat.Available, s)
	} else {
		stat.Occupied = append(stat.Occupied, s)
	}
}

func pingIP(start, end string) {

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
	var wg sync.WaitGroup
	for _, v := range hostslice {
		wg.Add(1)
		go reachcheck(v, &wg)
	}
	wg.Wait()
}
func checkIP(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	startIP := ps.ByName("start")
	endIP := ps.ByName("end")
	pingIP(startIP, endIP)

	type iplist struct {
		Available []net.IP
		Occupied  []net.IP
		AvailLen	int
	}
	ipl := iplist{}

	ipl.AvailLen = len(ipl.Available)
	ipl.Available = make([]net.IP, 0, len(stat.Available))
	for _, ip := range stat.Available {
		ipl.Available = append(ipl.Available, net.ParseIP(ip))
	}
	sort.Slice(ipl.Available, func(i, j int) bool{
		return bytes.Compare(ipl.Available[i], ipl.Available[j]) <0
	})

	ipl.Occupied = make([]net.IP, 0, len(stat.Occupied))
	for _, ip := range stat.Occupied {
		ipl.Occupied = append(ipl.Occupied, net.ParseIP(ip))
	}
	sort.Slice(ipl.Occupied, func(i, j int) bool{
		return bytes.Compare(ipl.Occupied[i], ipl.Occupied[j]) <0
	})

	fmt.Println(ipl)
	fp := path.Join("templates", "index.html")
	tmpl, err := template.ParseFiles(fp)
	if err != nil {
		fmt.Println(err)
		//http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := tmpl.Execute(w, ipl); err != nil {
		fmt.Println(err)
		//http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func main() {
	router := httprouter.New()
	router.GET("/:start/:end", checkIP)
	log.Fatal(http.ListenAndServe(":5000", router))
}
