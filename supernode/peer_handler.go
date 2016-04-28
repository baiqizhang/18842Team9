package main

import (
	"bufio"
	"dsproject/util"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net"
	"strconv"
	"strings"
	"time"
)

var lastClient *util.Client // = nil
var heartConn net.Conn
var normalConn net.Conn
var failtoken util.FailureToken

func reestablishRing(token util.FailureToken) {
	fmt.Println("My neighbor has failed")
	peerAddr := token.InitiatedNode
	fmt.Println("peerAddr: " + peerAddr)
	processCommand("PEERADDR " + peerAddr)
}

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
	heartport := nextNextPort(port)
	fmt.Println("Supernode Listening at " + heartport + " for Heartbeat connection")
	listener, err := net.Listen("tcp", ":"+heartport)
	util.CheckError(err)
	/* 1 means first conn, 0 means it's not*/
	firstConn := 1

	var heart int
	var heartbeatFrom string

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
						/* Generate correct format of the address of failed node so that the other nodes can detect */
						failAddr := lastClient.Conn.RemoteAddr().String()
						getIP := strings.Split(failAddr, ":")
						newFailAddr := getIP[0] + ":" + heartbeatFrom
						fmt.Println("Failed Address is " + newFailAddr)

						/* Generate the correct format for the address of initiated node so that other nodes can connect */
						localAddr := normalConn.LocalAddr().String()
						connectToPort := strings.Split(localAddr, ":")
						newLocalAddr := connectToPort[0] + ":" + port
						fmt.Println("Failure Initiated by node " + newLocalAddr)

						/* Creating a fail token */
						failtoken.FailAddr = newFailAddr
						failtoken.InitiatedNode = newLocalAddr
						failtokenByte, _ := json.Marshal(failtoken)

						remoteAddr := normalConn.RemoteAddr().String()
						failed := failtoken.FailAddr
						if failed == remoteAddr {
							reestablishRing(failtoken)
						} else {
							failtokenStr := string(failtokenByte)
							failureWriter := bufio.NewWriter(normalConn)
							failureWriter.WriteString("FAILURE " + failtokenStr + "\n")
							failureWriter.Flush()
						}

						/* Prepare for the next new connection and get the heartbeat */
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
			/* Get the port to set in case of failure handling */
			msgSplit := strings.Split(strings.Trim(msg, "\r\n"), " ")
			heartbeatFrom = msgSplit[2]
			if util.Verbose == 1 {
				fmt.Print("[listenHeart] " + time.Now().Format("20060102150405") + " " + msg + heartbeatFrom)
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
		if len(message) < 20 {
			fmt.Println("[handlePeer] received:" + message)
		}
		words := strings.Split(strings.Trim(message, "\r\n"), " ")

		// new connection
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
			fmt.Println("Received failure token")
			var token util.FailureToken
			err := json.Unmarshal([]byte(words[1]), &token)
			if err != nil {
				fmt.Println("error when unmarshalling failure token")
				continue
			}

			fmt.Println(token.FailAddr)
			fmt.Println(token.InitiatedNode)
			failed := token.FailAddr
			fmt.Println("Remote address of normal connection is " + normalConn.RemoteAddr().String())
			/*check if failed node is my neighbor*/
			if failed == normalConn.RemoteAddr().String() {
				fmt.Println("My neighbor has failed")
				/* establish a connection to the initiated node */
				reestablishRing(token)
			} else {
				/* forward the token as it is*/
				failureWriter := bufio.NewWriter(normalConn)
				failureWriter.WriteString("FAILURE " + words[1] + "\n")
				failureWriter.Flush()
			}
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
			len := token.Length
			isVisualization := false
			if len != 1 {
				isVisualization = true
			}
			//dest := util.ParseFloatCoordinates(args[3], args[4])
			for carNodeAddr, position := range idleCarNodePosition {
				if !isVisualization {
					fmt.Print("[PICKUP] CNAddr: " + carNodeAddr + " pos:")
					fmt.Print(position.X)
					fmt.Print(" ")
					fmt.Print(position.Y)
					fmt.Print(" dist: ")
				}
				dist := position.DistanceTo(source)
				doneFlag := false
				for i := 0; i < len; i++ {
					if token.Points[i].X == math.MaxFloat64/10 {
						token.Points[i] = position
						token.Addrs[i] = lastClient.Conn.LocalAddr().String() + "|" + carNodeAddr
						doneFlag = true
						break
					}
				}
				if !doneFlag {
					for i := 0; i < len; i++ {
						if dist < token.Points[i].DistanceTo(source) {
							token.Points[i] = position
							token.Addrs[i] = lastClient.Conn.LocalAddr().String() + "|" + carNodeAddr
							break
						}
					}
				}
			}
			if !isVisualization {
				fmt.Print("[PICKUP] local result: " + token.Addrs[0] + " = ")
				fmt.Print(token.Points[0].X)
				fmt.Print(" ")
				fmt.Println(token.Points[0].X)
			}
			// check if we've went throught the ring
			origin := lastClient.Conn.LocalAddr().String()
			if origin == token.Origin {
				if !isVisualization {
					finalAddr := token.Addrs[0]
					if !isVisualization {
						fmt.Print("[PICKUP] FINAL RESULT: " + finalAddr + " = ")
						fmt.Print(token.Points[0].X)
						fmt.Print(" ")
						fmt.Println(token.Points[0].X)
					}
					addrs := strings.Split(strings.Trim(finalAddr, "\r\n"), "|")
					snAddr := addrs[0]
					cnAddr := addrs[1]
					if !isVisualization {
						fmt.Println("[PICKUP] inform SN: " + snAddr + "-> CN:" + cnAddr)
					}
					tcpAddr, err := net.ResolveTCPAddr("tcp4", snAddr)
					util.CheckError(err)
					destSNAddr, err := net.DialTCP("tcp", nil, tcpAddr)
					util.CheckError(err)

					destSNWriter := bufio.NewWriter(destSNAddr)
					destSNWriter.WriteString("PICKUP " + cnAddr + " ")
					destSNWriter.WriteString(strconv.FormatFloat(token.Src.X, 'f', 4, 64) + " ")
					destSNWriter.WriteString(strconv.FormatFloat(token.Src.Y, 'f', 4, 64) + " ")
					destSNWriter.WriteString(strconv.FormatFloat(token.Dest.X, 'f', 4, 64) + " ")
					destSNWriter.WriteString(strconv.FormatFloat(token.Dest.Y, 'f', 4, 64) + "\n")
					destSNWriter.Flush()

					destSNReader := bufio.NewReader(destSNAddr)
					response, _ := destSNReader.ReadString('\n')
					response = strings.Trim(response, "\r\n")
					if !isVisualization {
						fmt.Println("[PICKUP] response received:" + response)
					}
					reqMap[token.ReqID] = response
				} else {
					tokenByte, _ := json.Marshal(token)
					tokenStr := string(tokenByte)
					visStr = tokenStr
					// fmt.Println("[UI] Points:" + tokenStr)
				}
			} else {
				tokenByte, _ := json.Marshal(token)
				tokenStr := string(tokenByte)
				if !isVisualization {
					fmt.Println("[PICKUP] Pass token: " + tokenStr)
				}
				writerToNextNode := bufio.NewWriter(normalConn)
				writerToNextNode.WriteString("PICKUP_TOKEN " + tokenStr + "\n")
				writerToNextNode.Flush()
			}
		}

		// some SN ask this SN to inform some CN to pickup
		if words[0] == "PICKUP" {
			cnAddr := words[1]
			for addr, conn := range carNodeConn {
				if addr == cnAddr {
					fmt.Println("[PICKUP] inform CN:" + cnAddr)
					cnWriter := bufio.NewWriter(conn)
					_, err := cnWriter.WriteString("PICKUP " + words[2] + " " + words[3] + " " + words[4] + " " + words[5] + "\n")
					cnWriter.Flush()

					snWriter := bufio.NewWriter(client.Conn)
					if err != nil {
						snWriter.WriteString("PICKUP FAIL\n")
					} else {
						snWriter.WriteString("PICKUP OK\n")
					}

					snWriter.Flush()
				}
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
	go dialHeart(ip + ":" + nextNextPort(dstport))
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

func nextNextPort(portstr string) string {
	intPort, _ := strconv.Atoi(portstr)
	nnport := strconv.Itoa(intPort + 2)
	return nnport
}
