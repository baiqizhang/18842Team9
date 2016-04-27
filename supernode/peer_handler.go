package main

import (
	"bufio"
	"dsproject/util"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"time"
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

/* handle the heartbeat connection while listening to it*/
func listenHeart() {
	heartport := next_next_port(port)
	fmt.Println("Supernode Listening at " + heartport + " for Heartbeat connection")
	listener, err := net.Listen("tcp", ":"+heartport)
	util.CheckError(err)
    /* 1 means first conn, 0 means it's not*/
    firstConn := 1

    var heart int

	for {
		conn, err := listener.Accept()
		util.CheckError(err)
        /* As soon as it gets the first connection, heartbeat is alive */
        heart = 1
        /* Start the thread only once when a peer joins because every node can have only one peer*/
        if firstConn == 1 {
            firstConn = 0
            go func() {
		        for {
				    time.Sleep(5000 * time.Millisecond)
				    if heart == 0 {
					    fmt.Println("Start Failure handling")
                        localAddr := normalConn.LocalAddr().String()
                        fmt.Println("Local IP is " +localAddr)
                        failAddr := lastClient.Conn.RemoteAddr().String()
                        fmt.Println("Failed node's IP is " +failAddr)

                        failureWriter := bufio.NewWriter(normalConn)
                        failureWriter.WriteString("FAILURE\n")
                        failureWriter.Flush()


                        firstConn = 1
					    break
			    	}
                    /* Reset heartbeat to 0 to count the next heartbeat*/
				    heart = 0
			    }
            }()
        }

		reader := bufio.NewReader(conn)
		for {
			msg, err := reader.ReadString('\n')
			if err != nil {
                conn.Close()
				break
			}
			if util.Verbose == 1 {
				fmt.Print("[listenHeart] " + time.Now().Format("20060102150405") + " " + msg)
			}
            /* Set heartbeat to 1*/
			heart = 1
		}
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
		fmt.Println("[handlePeer] received:" + message)
		words := strings.Split(strings.Trim(message, "\r\n"), " ")
		if words[0] == "NEWCONN" {
			remoteListeningPort := words[1]
			fmt.Println("[handlePeer] new connection from " + client.Conn.RemoteAddr().String() + "(" + remoteListeningPort + ")")
			if lastClient != nil {
				writer := bufio.NewWriter(lastClient.Conn)
				newPeerAddr := client.Conn.RemoteAddr()
				writer.WriteString("REDIRECT " + newPeerAddr.(*net.TCPAddr).IP.String() + ":" + words[1] + "\n")
				writer.Flush()
			}
			lastClient = &client
		}

        /* To handle failure */
        if words[0] == "FAILURE" {
        fmt.Println("Received a failure handling message")
            /*check if failed node is my neighbor*/

            /* forward the token as it is*/
        }
		// get a PICKUP_TOKEN, update the result
		if words[0] == "PICKUP_TOKEN" {
			var token util.PickupToken
			err := json.Unmarshal([]byte(words[1]), &token)
			if err != nil {
				fmt.Println("error when unmarshalling message")
				continue
			}

			source := token.Src
			finalResult := token.Points[0].DistanceTo(source)
			finalPoint := token.Points[0]
			finalAddr := token.Addrs[0]
			//dest := util.ParseFloatCoordinates(args[3], args[4])
			for carNodeAddr, position := range idleCarNodePosition {
				fmt.Print("[PICKUP] CNAddr: " + carNodeAddr + " pos:")
				fmt.Print(position.X)
				fmt.Print(" ")
				fmt.Print(position.Y)
				fmt.Print(" dist: ")
				dist := position.DistanceTo(source)
				if dist < finalResult {
					finalResult = dist
					finalPoint = position
					finalAddr = lastClient.Conn.LocalAddr().String() + "|" + carNodeAddr
				}
				fmt.Println(dist)
			}
			token.Points[0] = finalPoint
			token.Addrs[0] = finalAddr

			fmt.Print("[PICKUP] local result: " + finalAddr + " = ")
			fmt.Println(finalResult)

			// check if we've went throught the ring
			origin := lastClient.Conn.LocalAddr().String()
			if origin == token.Origin {
				fmt.Print("[PICKUP] FINAL RESULT: " + finalAddr + " = ")
				fmt.Println(finalResult)
			} else {
				tokenByte, _ := json.Marshal(token)
				tokenStr := string(tokenByte)
				fmt.Println("[PICKUP] Pass token: " + tokenStr)

				writerToNextNode := bufio.NewWriter(normalConn)
				writerToNextNode.WriteString("PICKUP_TOKEN " + tokenStr + "\n")
				writerToNextNode.Flush()
			}
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
		writer.WriteString("HEARTBEAT from " + port + "\n")
		writer.Flush()
		time.Sleep(1000 * time.Millisecond)
	}
}

func next_next_port(portstr string) string {
	intPort, _ := strconv.Atoi(portstr)
	nnport := strconv.Itoa(intPort + 2)
	return nnport
}
