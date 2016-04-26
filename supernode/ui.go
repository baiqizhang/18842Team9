// Package visualization implements the data structure of the Google
// Visualization API.
package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// Default HTTP Request Handler for UI
func dataHandler(w http.ResponseWriter, r *http.Request) {
	d := DataTable{
		ColsDesc: []ColDesc{
			{Label: "X", Type: "number"},
			{Label: "Y", Type: "number"},
			{Label: "Y", Type: "number"},
		},
		Rows: []Row{
			{
				C: []ColVal{
					{
						V: 4,
					},
					{
						V: 3,
					},
					{
						V: "null",
					},
				},
			},
			{
				C: []ColVal{
					{
						V: -1,
					},
					{
						V: "null",
					},
					{
						V: -7,
					},
				},
			},
		},
	}
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
