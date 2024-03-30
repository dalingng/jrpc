package jrpc_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"testing"

	"github.com/dalingng/jrpc"
)

type Main struct {
}


// 嵌套方法
func (m *Main)ChildrenMethods() ([]any, error){
    return []any{
        &Info{},
    },nil
}

func (m *Main) Test(ctx context.Context, req struct {
	Nickname *string
}) (string, error) {
	if req.Nickname == nil {
		return "", jrpc.NewError(10, "请填写用户名", nil)
	}
	return "你好:" + *req.Nickname, nil
}

type Info struct{
}
func (m *Info) Get(ctx context.Context)(string, error){
    return "jrpc简单示例",nil
}

func Test(t *testing.T) {

	jsonrpc := jrpc.JSONRPC{}
	jsonrpc.Register(&Main{})

	req := `[{"id":1,"jsonrpc":"2.0","method":"Main.Test","params":{}},{"id":2,"jsonrpc":"2.0","method":"Main.Test","params":{"Nickname":"大宁"}},{"id":3,"jsonrpc":"2.0","method":"Main.Info.Get","params":null}]`
	t.Log("=========请求================")
	t.Log("\n" + string(req))

	res := jsonrpc.Call(context.Background(), []byte(req))
	t.Log("=========返回结果================")
	t.Log("\n" + string(res))

	start := make(chan bool)
	go func() {
		listen, err := net.Listen("tcp", ":7890")
		if err != nil {
			t.Fatal(err.Error())
		}
		start <- true
		for {
			conn, err := listen.Accept()
			if err != nil {
				t.Fatal(err.Error())
			}
			enc := json.NewEncoder(conn)
			dec := json.NewDecoder(conn)
			for {
				data := json.RawMessage{}
				err := dec.Decode(&data)
				if err != nil {
					t.Fatal(err.Error())
				}
				res := jsonrpc.Call(context.Background(), data)
				enc.Encode(json.RawMessage(res))
			}
		}
	}()
	// 等待socket服务器启动
	<-start

	t.Log("=========socket发送================")
	conn, err := net.Dial("tcp", "127.0.0.1:7890")
	if err != nil {
		t.Fatal(err.Error())
	}
	enc := json.NewEncoder(conn)
	dec := json.NewDecoder(conn)

	reqData := jrpc.Request{
		Id:      json.RawMessage("1"),
		Jsonrpc: json.RawMessage("2.0"),
		Method:  "Main.Test",
		Params:  json.RawMessage("null"),
	}
	t.Log("=========请求================")
	t.Log(fmt.Sprintf("%+v", reqData))
	enc.Encode(reqData)

	r := json.RawMessage{}
	dec.Decode(&r)
	t.Log("=========返回结果================")
	t.Log("\n" + string(r))
    

}
