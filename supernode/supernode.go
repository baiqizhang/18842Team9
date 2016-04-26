package main

import (
	"dsproject/util"
	"fmt"
	"net/http"
	"os"
	"strconv"
)

var port string
var clients []util.Client
var name string

//REQMAP map for <request id, request struct>
var REQMAP = make(map[string]util.Request)

//COUNTCAR counter for carnodes which are ordinary nodes and counter for supernodes
var COUNTCAR int // 0 is the default value

//COUNTSUPER variable export comment placeholder
var COUNTSUPER int // 0 is the default value

func main() {
	args := os.Args[1:]
	if len(args) < 1 {
		fmt.Println("Usage: supernode PORT [-v]")
		os.Exit(0)
	}
	port = args[0]
	//name = args[2]
	for _, arg := range args {
		if arg == "-v" {
			util.Verbose = 1
		}
	}

	// connect to frontend instance
	go dialServer()

	// listen to peer(SuperNode) connection in Ring Topology
	go listenPeer()
	go listenHeart()

	// a UI
	go listenHTTP()
	// listen to node connection requests? (not sure if is required)
	listenCarNode()
}

func listenHTTP() {
	intPort, _ := strconv.Atoi(port)
	httpport := strconv.Itoa(intPort + 3)

	fmt.Print("web server running on " + httpport + "\n")
	http.Handle("/", http.FileServer(http.Dir("../server/public")))
	http.HandleFunc("/api/data", dataHandler)
	http.HandleFunc("/api/ride", rideHandler)
	http.ListenAndServe(":"+httpport, nil)
}

// Default HTTP Request Handler for UI
func rideHandler(w http.ResponseWriter, r *http.Request) {
	sx := r.URL.Query().Get("sx")
	sy := r.URL.Query().Get("sy")
	dx := r.URL.Query().Get("dx")
	dy := r.URL.Query().Get("dy")
	fmt.Println("[UI] ride request received: " + sx + " " + sy + " " + dx + " " + dy)
	fmt.Fprintf(w, "result unknown\n")
}
