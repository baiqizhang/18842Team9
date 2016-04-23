package main

import (
	"dsproject/util"
	"fmt"
	"os"
)

var port string
var clients []util.Client

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
	for _, arg := range args {
		if arg == "-v" {
			util.Verbose = 1
		}
	}

	// connect to frontend instance
	go dialServer()

	// listen to peer(SuperNode) connection in Ring Topology
	go listenPeer()

	// listen to node connection requests? (not sure if is required)
	listenCarNode()
}
