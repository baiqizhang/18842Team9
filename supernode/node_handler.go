package main

import (
	"bufio"
	"dsproject/util"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
)

func listenCarNode() {
	carNodePortInt, err := strconv.Atoi(port)
	carNodePortString := strconv.Itoa(carNodePortInt + 1)

	listener, err := net.Listen("tcp", ":"+carNodePortString)
	util.CheckError(err)
	fmt.Println("Supernode Listening at " + carNodePortString + " for CarNode connection")
	for {
		conn, err := listener.Accept()
		util.CheckError(err)

		newClient := util.Client{Conn: conn, Name: "none"}
		clients = append(clients, newClient)
		go handleNode(newClient)

	}
}

func handleNode(client util.Client) {
	fmt.Println(client.Conn.RemoteAddr().String())
	reader := bufio.NewReader(client.Conn)

	message, _ := reader.ReadString('\n')

	// if connection comes from CarNode
	if message == "CARNODE\n" {
		fmt.Println("this is a " + message)
		COUNTCAR++
	}

	// Read handler
	for {
		message, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}
		util.CheckError(err)

		fmt.Println("[Node Message]:" + message)
		words := strings.Split(strings.Trim(message, "\r\n"), " ")

		// if connection comes from CarNode
		if words[0] == "COMPUTERESULT" {
			fmt.Println("[Node] Register Name:" + words[1])
			client.Name = words[1]
			result, _ := strconv.ParseFloat(words[2], 64)

			fmt.Println("id: " + words[5])
			fmt.Println("Carname: " + words[1] + " " + REQMAP[words[5]].Carname)
			fmt.Println("Count: " + strconv.Itoa(REQMAP[words[5]].Count))

			rq := REQMAP[words[5]]
			if result < REQMAP[words[5]].FinalResult {
				// finalResult = result
				// finalClient = client
				rq.FinalResult = result
				rq.FinalConn = client.Conn
				rq.Carname = words[1]
			}

			rq.Count--
			REQMAP[words[5]] = rq

			if REQMAP[words[5]].Count == 0 {
				fmt.Println("SEND REQ TO: " + REQMAP[words[5]].Carname)

				delete(REQMAP, words[5])

				writer := bufio.NewWriter(rq.FinalConn)
				writer.WriteString("PICKUP " + rq.Source + " " + rq.Dest + "\n")
				writer.Flush()
			}
		}

	}
}
