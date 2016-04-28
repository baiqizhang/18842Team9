package main

import (
	"bufio"
	"dsproject/util"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"os"
	"strconv"
)

var port string

var reqMap = make(map[int]*http.ResponseWriter)
var reqID int

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

	var token util.PickupToken
	token.ReqID = reqID
	token.Origin = lastClient.Conn.LocalAddr().String()
	token.Src = *util.ParseFloatCoordinates(sx, sy)
	token.Length = 1
	token.Points = make([]util.Point, 1)
	token.Points[0] = util.Point{X: math.MaxFloat64 / 10, Y: math.MaxFloat64 / 10}
	token.Addrs = make([]string, 1)

	tokenByte, _ := json.Marshal(token)
	tokenStr := string(tokenByte)
	fmt.Println("[rideHandler] send token: " + tokenStr)

	reqMap[reqID] = &w
	reqID++

	writerToNextNode := bufio.NewWriter(normalConn)
	writerToNextNode.WriteString("PICKUP_TOKEN " + tokenStr + "\n")
	writerToNextNode.Flush()
}
