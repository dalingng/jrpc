package jrpc

import "encoding/json"

func GetMethods(rpc JSONRPC, req *Request) []byte {
	methods := []string{}
	for _, v := range rpc {
		methods = append(methods, v.name)
	}
	resp := &Response{
		Id:      req.Id,
		Jsonrpc: req.Jsonrpc,
		Result:  methods,
	}
	b, _ := json.Marshal(resp)
	return b
}
