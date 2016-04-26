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

func dialServer() {
	tcpAddr, err := net.ResolveTCPAddr("tcp4", util.ServerAddr)
	util.CheckError(err)

	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	util.CheckError(err)

	_, err = conn.Write([]byte("SUPERNODE REGISTER " + port + "\r\n"))
	util.CheckError(err)

	// send Heartbeat
	go func() {
		for {
			writer := bufio.NewWriter(conn)
			writer.WriteString("SUPERNODE HEARTBEAT " + port + "\n")
			writer.Flush()
			time.Sleep(1000 * time.Millisecond)
		}
	}()

	reader := bufio.NewReader(conn)
	// Read handler
	for {
		message, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}
		util.CheckError(err)

		fmt.Println("[Server Message]:" + message)
		processCommand(message)
	}
}

// process command from Server
func processCommand(cmd string) {
	args := strings.Split(strings.Trim(cmd, "\r\n"), " ")

	//Compute distance to the customer
	if args[0] == "PEERADDR" {
		peerAddr := args[1]
		words := strings.Split(peerAddr, ":")
		go dialPeer(peerAddr)
		go dialHeart(words[0] + ":" + next_next_port(words[1]))
	}

	//Compute distance to the customer
	if args[0] == "PICKUP" {
		/*
			finalResult := math.MaxFloat64
			var finalAddr string
			source := util.ParseFloatCoordinates(args[1], args[2])
			//dest := util.ParseFloatCoordinates(args[3], args[4])
			for carNodeAddr, position := range idleCarNodePosition {
				fmt.Print("[PICKUP] CNAddr: " + carNodeAddr + " pos:")
				fmt.Print(position.X)
				fmt.Print(" ")
				fmt.Print(position.Y)
				fmt.Print(" dist: ")
				dist := position.DistanceTo(*source)
				if dist < finalResult {
					finalResult = dist
					finalAddr = carNodeAddr
				}
				fmt.Println(dist)
			}
			fmt.Print("[PICKUP] final result: " + finalAddr + " = ")
			fmt.Println(finalResult)

			// tell the CarNode to pickup
			writer := bufio.NewWriter(carNodeConn[finalAddr])
			writer.WriteString("PICKUP " + args[1] + " " + args[2] + " " + args[3] + " " + args[4] + "\n")
			writer.Flush()
		*/
		writerToNextNode := bufio.NewWriter(normalConn)
		writerToNextNode.WriteString("PICKUP_TOKEN " + normalConn.LocalAddr().String() + ":" + port)
		writerToNextNode.WriteString(" " + args[1] + " " + args[2] + " ")
		writerToNextNode.WriteString("\n")
		writerToNextNode.Flush()
	}
}
