package main

import (
	"bufio"
	"dsproject/util"
	"fmt"
	"io"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

var lastSNClient *util.Client // = nil

var superNodeAliveCounter = make(map[string]int)
var aliveSuperNodeAddrs []string
var l sync.Mutex

//check if connection from SN is broken
func checkConnection() {
	for {
		l.Lock()
		aliveSuperNodeAddrs = make([]string, 0, len(superNodeAliveCounter))
		for k := range superNodeAliveCounter {
			aliveSuperNodeAddrs = append(aliveSuperNodeAddrs, k)
		}

		for key, value := range superNodeAliveCounter {
			superNodeAliveCounter[key] = value - 1
			if util.Verbose == 1 {
				fmt.Printf("Heartbeat count %s = %d\r\n", key, value)
			}
			if value == 0 {
				delete(superNodeAliveCounter, key)
				fmt.Println("[Supernode] Lost connection: " + key)
			}
		}
		l.Unlock()
		time.Sleep(500 * time.Millisecond)
	}
}

//handle message from a SN or CarNode
func handleClient(client *util.Client) {
	conn := client.Conn
	fmt.Println("[New Client]:" + conn.RemoteAddr().String())

	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)

	// Read handler
	for {
		message, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}
		util.CheckError(err)

		if util.Verbose == 1 {
			fmt.Print("[Message Received]:" + message)
		}

		words := strings.Split(strings.Trim(message, "\r\n"), " ")

		// if connection comes from CarNode
		client.Type = words[0]
		if words[0] == "SUPERNODE" {
			newSNAddr := client.Conn.RemoteAddr()
			newSNIPStr := newSNAddr.(*net.TCPAddr).IP.String()
			newSNPortStr := words[2]
			//mark the client as alive
			superNodeAliveCounter[newSNIPStr+":"+newSNPortStr] = 5

			// check the 2nd arg
			if words[1] == "REGISTER" {
				fmt.Print("[SuperNode]" + message)
				client.Name = newSNPortStr
				if lastSNClient != nil {
					//tell the newcomer the last supernode's ip:port
					lastSNAddr := lastSNClient.Conn.RemoteAddr()
					writer.WriteString("PEERADDR " + lastSNAddr.(*net.TCPAddr).IP.String() + ":" + lastSNClient.Name + "\n")
					writer.Flush()

					// close the ring when node count = 2
					fmt.Println(superNodeAliveCounter)
					if len(superNodeAliveCounter) == 2 {
						tempWriter := bufio.NewWriter(lastSNClient.Conn)
						tempWriter.WriteString("PEERADDR " + newSNIPStr + ":" + newSNPortStr + "\n")
						tempWriter.Flush()
					}
				}
				lastSNClient = client
				continue
			}
			if words[1] == "HEARTBEAT" {
				continue
			}
		}
		if words[0] == "NODE" {
			if words[1] == "REGISTER" {
				for {
					l.Lock()
					aliveSuperNodeAddrs = make([]string, 0, len(superNodeAliveCounter))
					for k := range superNodeAliveCounter {
						aliveSuperNodeAddrs = append(aliveSuperNodeAddrs, k)
					}
					fmt.Println(aliveSuperNodeAddrs)
					fmt.Println(len(aliveSuperNodeAddrs))
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
				SNPort = strconv.Itoa(SNPortInt + 1)

				writer.WriteString(SNIP + ":" + SNPort + "\n")
				writer.Flush()
				client.Conn.Close()
				return
			}
		}
		fmt.Println("error: message not recognized")
	}
}
