// Package visualization implements the data structure of the Google
// Visualization API.
package main

import (
	"bufio"
	"dsproject/util"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"time"
)

var visStr string

func getVis() {
	time.Sleep(2000 * time.Millisecond)
	for {
		var token util.PickupToken
		token.ReqID = -1
		token.Origin = lastClient.Conn.LocalAddr().String()
		token.Src = util.Point{X: 0, Y: 0}
		token.Length = 10
		token.Points = make([]util.Point, 10)
		for i := 0; i < token.Length; i++ {
			token.Points[i] = util.Point{X: math.MaxFloat64 / 10, Y: math.MaxFloat64 / 10}
		}
		token.Addrs = make([]string, 10)

		tokenByte, _ := json.Marshal(token)
		tokenStr := string(tokenByte)

		writerToNextNode := bufio.NewWriter(normalConn)
		writerToNextNode.WriteString("PICKUP_TOKEN " + tokenStr + "\n")
		writerToNextNode.Flush()

		time.Sleep(1000 * time.Millisecond)
	}
}

// Default HTTP Request Handler for UI
func dataHandler(w http.ResponseWriter, r *http.Request) {
	var token util.PickupToken
	err := json.Unmarshal([]byte(visStr), &token)
	if err != nil {
		fmt.Println("error when unmarshalling message")
		return
	}
	row := make([]Row, 0, 10)
	// fmt.Println("[UI] " + visStr)

	for i := 0; i < token.Length; i++ {
		point := token.Points[i]
		if point.X == math.MaxFloat64/10 {
			continue
		}
		row = append(row, Row{
			C: []ColVal{
				{
					V: point.X,
				},
				{
					V: point.Y,
				},
			},
		})
	}

	// //forced sync, terrible idea
	// for reqMap[id] == "" {
	// 	time.Sleep(10 * time.Millisecond)
	// }
	// fmt.Println("[UI] response sent:" + reqMap[id])
	// fmt.Fprintf(w, "%s\r\n", reqMap[id])
	// delete(reqMap, id)

	for _, point := range idleCarNodePosition {
		row = append(row, Row{
			C: []ColVal{
				{
					V: point.X,
				},
				{
					V: point.Y,
				},
			},
		})
	}
	d := DataTable{
		ColsDesc: []ColDesc{
			{Label: "X", Type: "number"},
			{Label: "Y", Type: "number"},
			//{Label: "Y", Type: "number"},
		},
		Rows: row,
	}
	// d := DataTable{
	// 	ColsDesc: []ColDesc{
	// 		{Label: "X", Type: "number"},
	// 		{Label: "Y", Type: "number"},
	// 		//{Label: "Y", Type: "number"},
	// 	},
	// 	Rows: []Row{
	// 		{
	// 			C: []ColVal{
	// 				{
	// 					V: 4,
	// 				},
	// 				{
	// 					V: 3,
	// 				},
	// 				{
	// 					V: "null",
	// 				},
	// 			},
	// 		},
	// 		{
	// 			C: []ColVal{
	// 				{
	// 					V: -1,
	// 				},
	// 				{
	// 					V: "null",
	// 				},
	// 				{
	// 					V: -7,
	// 				},
	// 			},
	// 		},
	// 	},
	// }
	b, err := json.MarshalIndent(d, "", "	")
	if err != nil {
		fmt.Println(err)
	}
	// fmt.Printf("%s\n", b)
	fmt.Fprintf(w, "%s\n", b)
	// fmt.Fprintf(w, "<h1>Hello from Team 9 %s!</h1>", r.URL.Path[1:])
}

// ColDesc represents a description of a column in the Google
// visualization API.
type ColDesc struct {
	ID      string                 `json:"id,omitempty"`
	Label   string                 `json:"label,omitempty"`
	Type    string                 `json:"type"`
	Pattern string                 `json:"pattern,omitempty"`
	P       map[string]interface{} `json:"p,omitempty"`
}

// ColVal represents the value for a column cell.
type ColVal struct {
	V interface{}            `json:"v,omitempty"`
	F string                 `json:"f,omitempty"`
	P map[string]interface{} `json:"p,omitempty"`
}

// Row represents a row of data in the table.
type Row struct {
	C []ColVal `json:"c"`
}

// DataTable represents a Google Visualization data object.
type DataTable struct {
	ColsDesc []ColDesc `json:"cols"`
	Rows     []Row     `json:"rows"`
}
