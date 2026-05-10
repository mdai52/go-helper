package cloudflare

import (
	"encoding/json"
)

type RequestParam struct {
	Token   string          `note:"Api Token" json:"token"`
	Method  string          `binding:"required" json:"method"`
	Path    string          `binding:"required" json:"path"`
	Query   string          `note:"请求参数" json:"query"`
	Payload json.RawMessage `json:"payload"`
}

type ResponseData struct {
	Success    bool
	Errors     []ResponseInfo
	Messages   []ResponseInfo
	Result     any
	ResultInfo ResultInfo
}

type ResponseInfo struct {
	Code    int
	Message string
}

type ResultInfo struct {
	Page       int
	PerPage    int
	TotalPages int
	Count      int
	Total      int
	Cursor     string
	Cursors    ResultInfoCursors
}

type ResultInfoCursors struct {
	Before string
	After  string
}

//// Output

type Output struct {
	Datasets any
	Messages string `json:",omitempty"`
}

type OutputInfo struct {
	Page    int
	PerPage int
	Total   int
}

type OutputWithInfo struct {
	*Output
	DataInfo *OutputInfo
}
