// Package visualization implements the data structure of the Google
// Visualization API.
package main

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
