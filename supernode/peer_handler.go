package main

import (
	"bufio"
	"dsproject/util"
	"fmt"
	"io"
	"net"
	"strings"
	"time"
)

var lastClient *util.Client // = nil

func listenPeer() {
	// listen to node connection requests? (not sure if is required)
	listener, err := net.Listen("tcp", ":"+port)
	util.CheckError(err)
	fmt.Println("Supernode Listening at " + port + " for SuperNode connection")
	for {
		conn, err := listener.Accept()
		util.CheckError(err)

		newClient := util.Client{Conn: conn, Name: "none"}

		// clients = append(clients, newClient)
		go handlePeer(newClient)
	}

}

func handlePeer(client util.Client) {
	fmt.Println("[Peer Listener] new connection from" + client.Conn.RemoteAddr().String())
	reader := bufio.NewReader(client.Conn)

	// Read message that comes from previous node(passive connection)
	for {
		message, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}
		util.CheckError(err)

		if util.Verbose == 1 {
			fmt.Println("[Previous Node Message]:" + message)
		}
		words := strings.Split(strings.Trim(message, "\r\n"), " ")

		if words[0] == "NEWCONN" {
			fmt.Println("[Previous Node Message]:" + message)
			if lastClient != nil {
				writer := bufio.NewWriter(lastClient.Conn)
				newPeerAddr := client.Conn.RemoteAddr()
				writer.WriteString("REDIRECT " + newPeerAddr.(*net.TCPAddr).IP.String() + ":" + words[1] + "\n")
				writer.Flush()
			}
			lastClient = &client
		}

	}
}

func dialPeer(peerAddr string) {
	tcpAddr, err := net.ResolveTCPAddr("tcp4", peerAddr)
	util.CheckError(err)

	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	util.CheckError(err)

	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)
	//1st message
	writer.WriteString("NEWCONN " + port + "\n")
	go func() {
		fmt.Println("[dialPeer]: sending HB to " + peerAddr)
		for {
			writer.WriteString("HEARTBEAT " + port + "\n")
			writer.Flush()
			time.Sleep(1000 * time.Millisecond)
		}
	}()

	// Read message that comes from the dialed node(active connection)
	for {
		message, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}
		util.CheckError(err)

		fmt.Println("[Next Peer Message]:" + message)

		words := strings.Split(strings.Trim(message, "\r\n"), " ")
		if words[0] == "REDIRECT" {
			conn.Close()
			dialPeer(words[1])
			break
		}
	}

}
