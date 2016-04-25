package main

import (
	"bufio"
	"dsproject/util"
	"fmt"
	"io"
	"net"
	"strings"
	"time"
	"strconv"
)

var lastClient *util.Client // = nil
var heartConn net.Conn
var normalConn net.Conn

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

func listenHeart() {
	heartport := next_next_port(port)
	fmt.Println("Supernode Listening at " + heartport + " for Heartbeat connection")
	listener, err := net.Listen("tcp", ":" + heartport)
	util.CheckError(err)

	heart := 1
	for {
		conn, err := listener.Accept()
		util.CheckError(err)

//		go func() {
			reader := bufio.NewReader(conn)
			for {
				msg, err := reader.ReadString('\n')
				if err != nil {
					conn.Close()
					break
				}
				fmt.Println(msg)
				heart = 1
			}
//		}()

		go func() {
			for {
				time.Sleep(5000 * time.Millisecond)
				if heart == 0 {
					fmt.Println("Start Failure handling")
					break
				}
				heart = 0
			}
		}()
	}
}

func handlePeer(client util.Client) {
	reader := bufio.NewReader(client.Conn)

	// Read message that comes from previous node(passive connection)
	for {
		message, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}
		util.CheckError(err)

		words := strings.Split(strings.Trim(message, "\r\n"), " ")
		if words[0] == "NEWCONN" {
			fmt.Println(time.Now().Format("20060102150405") + " new connection")
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

	normalConn, err = net.DialTCP("tcp", nil, tcpAddr)
	util.CheckError(err)

	reader := bufio.NewReader(normalConn)
	writer := bufio.NewWriter(normalConn)
	//1st message

	writer.WriteString("NEWCONN " + port + "\n")
	writer.Flush()

	// Read message that comes from the dialed node(active connection)
	for {
		message, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}
		util.CheckError(err)

		words := strings.Split(strings.Trim(message, "\r\n"), " ")
		if words[0] == "REDIRECT" {
			normalConn.Close()
			heartConn.Close()
			addr := strings.Split(words[1], ":")
			//fmt.Println(time.Now().Format("20060102150405") + " Being redirect to " + addr[0] + ":" + addr[1])
			handleRedirect(addr[0], addr[1])
			break
		}
	}
}

func handleRedirect(ip string, dstport string) {
	//fmt.Println(time.Now().Format("20060102150405") + " in handle redirect " + ip + " " + dstport)
	go dialPeer(ip + ":" + dstport)
	go dialHeart(ip + ":" + next_next_port(dstport))
}

func dialHeart(peerAddr string) {
	tcpAddr, err := net.ResolveTCPAddr("tcp4", peerAddr)
	util.CheckError(err)

	heartConn, err = net.DialTCP("tcp", nil, tcpAddr)
	util.CheckError(err)

	writer := bufio.NewWriter(heartConn)

	for {
		writer.WriteString("HEARTBEAT from " + name + "\n")
		writer.Flush()
		time.Sleep(1000 * time.Millisecond)
	}
}

func next_next_port(portstr string) string {
	intPort, _ := strconv.Atoi(portstr)
	nnport := strconv.Itoa(intPort + 2)
	return nnport
}
