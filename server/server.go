package main

import (
	"bufio"
	"dsproject/util"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// var clients []*util.Client
// var reqID int

func main() {
	args := os.Args[1:]
	for _, arg := range args {
		if arg == "-v" {
			util.Verbose = 1
		}
	}

	//start HTTP UI server at 8080
	go listenHTTP()

	// start TCP, listening at 7070
	go listenTCP()

	stdin := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("Enter command: ")
		stdin.ReadString('\n')
		// cmd, _ := stdin.ReadString('\n')
		// processCommand(cmd)
	}
}

// listen for SN/carN connection
func listenTCP() {
	listener, err := net.Listen("tcp", ":7070")
	util.CheckError(err)

	go checkConnection()

	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		newClient := util.Client{Conn: conn, Name: "none"}
		// clients = append(clients, &newClient)
		go handleClient(&newClient)
	}
}

/*
// process command from Server Terminal
func processCommand(cmd string) {
	args := strings.Split(strings.Trim(cmd, "\r\n"), " ")

	//Compute distance to the customer
	if args[0] == "PICKUP" {
		//Pickup the customer
		source := util.ParseFloatCoordinates(args[1], args[2])
		dest := util.ParseFloatCoordinates(args[3], args[4])
		if source == nil || dest == nil {
			fmt.Println("Error: incorrect PICKUP format:" + cmd)
			return
		}

		var zero []byte
		for _, client := range clients {
			if client.Type == "NODE" {
				continue
			}
			fmt.Println("client " + client.Type + " " + client.Name)
			conn := client.Conn
			fmt.Println(conn.RemoteAddr().String())
			reader := bufio.NewReader(conn)
			_, err := reader.Read(zero)
			if err != nil {
				continue
			}

			fmt.Println("[PICKUP] send to SN:" + client.Conn.RemoteAddr().String())
			writer := bufio.NewWriter(conn)
			writer.WriteString("PICKUP " + args[1] + " " + args[2] + " " + args[3] + " " + args[4] + " " + strconv.Itoa(reqID) + "\n")
			reqID++
			writer.Flush()
		}
	}
}
*/

func redirect(w http.ResponseWriter, r *http.Request) {
	for {
		l.Lock()
		aliveSuperNodeAddrs = make([]string, 0, len(superNodeAliveCounter))
		for k := range superNodeAliveCounter {
			aliveSuperNodeAddrs = append(aliveSuperNodeAddrs, k)
		}
		if len(aliveSuperNodeAddrs) == 0 {
			l.Unlock()
			fmt.Println("[Node Register] no SN available, waiting...")
			time.Sleep(1000 * time.Millisecond)
		} else {
			break
		}
	}
	fmt.Println("[Node Register] send a random supernode addr")
	index := rand.Intn(len(aliveSuperNodeAddrs))
	addrString := aliveSuperNodeAddrs[index]
	l.Unlock()

	// port for carnodes is port for SN + 1
	parts := strings.Split(addrString, ":")
	SNIP := parts[0]
	SNPort := parts[1]
	SNPortInt, _ := strconv.Atoi(SNPort)
	SNPort = strconv.Itoa(SNPortInt + 3)

	redirectAddr := "http://" + SNIP + ":" + SNPort
	fmt.Println("[UI] redirect: " + redirectAddr)
	http.Redirect(w, r, redirectAddr, 301)
}

func listenHTTP() {
	// http.Handle("/ride/", http.StripPrefix("/ride/", http.FileServer(http.Dir("../server/public"))))
	// http.HandleFunc("/api/data", dataHandler)
	fmt.Print("web server running on 8080\n")
	http.HandleFunc("/", redirect)
	http.ListenAndServe(":8080", nil)
}
