package main

import (
	"bytes"
	"flag"
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
	if (startIPLastOctet >= 255) || (endIPLastOctet >= 255) {
		fmt.Println("Wrong IP Range")
		return
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

	if len(stat.Occupied) <= 1 || len(stat.Available) <=1{
		fmt.Fprintf(w, "Check Your IP Range," +
			" It works only on last octet, Your Range is StartIP : %s", startIP)
	}

	type iplist struct {
		Available []net.IP
		Occupied  []net.IP
	}
	ipl := iplist{}

	ipl.Available = make([]net.IP, 0, len(stat.Available))
	for _, ip := range stat.Available {
		ipl.Available = append(ipl.Available, net.ParseIP(ip))
	}
	sort.Slice(ipl.Available, func(i, j int) bool {
		return bytes.Compare(ipl.Available[i], ipl.Available[j]) < 0
	})

	ipl.Occupied = make([]net.IP, 0, len(stat.Occupied))
	for _, ip := range stat.Occupied {
		ipl.Occupied = append(ipl.Occupied, net.ParseIP(ip))
	}
	sort.Slice(ipl.Occupied, func(i, j int) bool {
		return bytes.Compare(ipl.Occupied[i], ipl.Occupied[j]) < 0
	})

	fmt.Println(ipl)
	fp := path.Join("templates", "index.html")
	tmpl, err := template.ParseFiles(fp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := tmpl.Execute(w, ipl); err != nil {
		fmt.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func ipInput(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	err := r.ParseForm()
	if err != nil {
		fmt.Println(err)
	}
	t, err := template.ParseFiles("templates/submit.html")
	if err != nil {
		fmt.Println(err)
	}
	t.Execute(w, nil)
	fmt.Println("Start IP", r.Form["startIP"])
	fmt.Println("End IP", r.Form["endIP"])
	startIP := r.Form["startIP"]
	endIP := r.Form["endIP"]
	localhostres := fmt.Sprintf("/:%s/:%s", startIP, endIP)
	fmt.Println(localhostres)
}

var portflag *string
func main() {
	portflag = flag.String("port", "5000", "Enter the PortNumber for this service to run")
	flag.Parse()
	fmt.Printf("IP Ping Range Running in: %s", *portflag)
	router := httprouter.New()
	router.GET("/:start/:end", checkIP)
	router.GET("/", ipInput)
	log.Fatal(http.ListenAndServe(":5555", router))
}
