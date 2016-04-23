package main

import (
	"bufio"
	"dsproject/util"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
)

var clients []*util.Client
var reqID int

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
		cmd, _ := stdin.ReadString('\n')
		processCommand(cmd)
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
		clients = append(clients, &newClient)
		go handleClient(&newClient)
	}
}

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

func listenHTTP() {
	http.Handle("/ride/", http.StripPrefix("/ride/", http.FileServer(http.Dir("./public"))))
	http.HandleFunc("/api/data", dataHandler)
	http.ListenAndServe(":8080", nil)
	fmt.Print("web server running on 8080\n")
}

// Default HTTP Request Handler for UI
func dataHandler(w http.ResponseWriter, r *http.Request) {
	d := DataTable{
		ColsDesc: []ColDesc{
			{Label: "X", Type: "number"},
			{Label: "Y", Type: "number"},
			{Label: "Y", Type: "number"},
		},
		Rows: []Row{
			{
				C: []ColVal{
					{
						V: 4,
					},
					{
						V: 3,
					},
					{
						V: "null",
					},
				},
			},
			{
				C: []ColVal{
					{
						V: -1,
					},
					{
						V: "null",
					},
					{
						V: -7,
					},
				},
			},
		},
	}
	b, err := json.MarshalIndent(d, "", "	")
	if err != nil {
		fmt.Println(err)
	}
	// fmt.Printf("%s\n", b)
	fmt.Fprintf(w, "%s\n", b)
	// fmt.Fprintf(w, "<h1>Hello from Team 9 %s!</h1>", r.URL.Path[1:])
}
