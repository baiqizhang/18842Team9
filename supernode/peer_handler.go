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
        /* Accept socket connection and create a new client structure */
		conn, err := listener.Accept()
		util.CheckError(err)

		newClient := util.Client{Conn: conn}
        fmt.Println("[Incoming message local and remote: ]" +newClient.Conn.LocalAddr().String() + " " +newClient.Conn.RemoteAddr().String())

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
			fmt.Println("[Previous Peer Message]:" + message)
		}
		words := strings.Split(strings.Trim(message, "\r\n"), " ")

		if words[0] == "NEWCONN" {
			fmt.Println("[Previous Peer Message]:" + message)
			if lastClient != nil {
                fmt.Println("Last Peer before redirection is:")
                fmt.Println(lastClient.Conn.RemoteAddr().String())
				writer := bufio.NewWriter(lastClient.Conn)
                /* IP of previosuly established connection */
				newPeerAddr := client.Conn.RemoteAddr()
                fmt.Println("redirect this new connection " +newPeerAddr.(*net.TCPAddr).IP.String() + ":" + words[1])
				writer.WriteString("REDIRECT " + newPeerAddr.(*net.TCPAddr).IP.String() + ":" + words[1] + "\n")
				writer.Flush()
			}
            /* The other end from which the connection comes */
			lastClient = &client
            fmt.Println("Last client now is : ")
            fmt.Println(lastClient.Conn.RemoteAddr().String())
		}

	}
}

/*Make a connection with other SN peers by sending NEWCONN*/
func dialPeer(peerAddr string) {
	tcpAddr, err := net.ResolveTCPAddr("tcp4", peerAddr)
	util.CheckError(err)

	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	util.CheckError(err)

	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)
	//1st message, port of the SN
	writer.WriteString("NEWCONN " + port  + "\n")
	go func() {
		fmt.Println("[dialPeer]: sending HB to " + peerAddr)
		for {
			_, err := writer.WriteString("HEARTBEAT " + port + "\n")
			writer.Flush()
			time.Sleep(1000 * time.Millisecond)
			if err != nil {
				break
			}
		}
		fmt.Println("[dialPeer] HB to " + peerAddr + " failed")
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
            fmt.Println("redirect this connection " +words[1])
			conn.Close()
			dialPeer(words[1])
			break
		}
	}

}
