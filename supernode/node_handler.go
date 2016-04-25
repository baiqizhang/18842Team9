package main

import (
	"bufio"
	"dsproject/util"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"time"
)

var idleCarNodePosition = make(map[string]util.Point)
var carNodeConn = make(map[string]net.Conn)

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
		go handleNode(newClient)

	}
}

func handleNode(client util.Client) {
	addrString := client.Conn.RemoteAddr().String()
	fmt.Println("[handleNode] New CarNode:" + addrString)

	reader := bufio.NewReader(client.Conn)
	writer := bufio.NewWriter(client.Conn)

	go func() {
		for {
			writer.WriteString("HEARTBEAT\n")
			writer.Flush()
			time.Sleep(2000 * time.Millisecond)
		}
	}()

	// Read handler
	for {
		message, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}
		util.CheckError(err)

		fmt.Println("[Node Message]:" + message)
		words := strings.Split(strings.Trim(message, "\r\n"), " ")

		if words[0] == "POSITION" {
			point := util.ParseFloatCoordinates(words[1], words[2])
			if words[3] == "IDLE" {
				idleCarNodePosition[addrString] = *point
				carNodeConn[addrString] = client.Conn
			} else {
				delete(idleCarNodePosition, addrString)
				delete(carNodeConn, addrString)
			}
			fmt.Println(idleCarNodePosition)
		}
	}
}
