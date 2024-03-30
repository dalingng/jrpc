关于jrpc
===
jrpc是一个非常简单JSONRPC实现

程序不考虑性能只考虑简单实现，实测性能还行

使用详情请看 jrpc_test.go

使用方法([]byte请求)
---

```
import (
	"context"
	"github.com/dalingng/jrpc"
)

type Main struct {
}

func (m *Main) Test(ctx context.Context, req struct {
	Nickname *string
}) (string, jrpc.Error) {
	if req.Nickname == nil {
		return "", jrpc.NewError(10, "请填写用户名", nil)
	}
	return "你好:" + *req.Nickname, nil
}
```

```
jsonrpc := jrpc.JSONRPC{}
// 注册方法
jsonrpc.Register(&Main{})

// 批量请求
req := `[{"id":1,"jsonrpc":"2.0","method":"Main.Test","params":{}},{"id":2,"jsonrpc":"2.0","method":"Main.Test","params":{"Nickname":"大宁"}}]`
res := jsonrpc.Call(context.Background(), []byte(req))
fmt.Println(string(res))

// 单个请求
req = `{"id":2,"jsonrpc":"2.0","method":"Main.Test","params":{"Nickname":"单个请求"}}`
res = jsonrpc.Call(context.Background(), []byte(req))
fmt.Println(string(res))
```
返回结果
```
[{"jsonrpc":"2.0","id":1,"result":"","error":{"code":10,"message":"请填写用户名"}},{"jsonrpc":"2.0","id":2,"result":"你好:大宁"}]
{"jsonrpc":"2.0","id":2,"result":"你好:单个请求"}
```

嵌套
===
需要实现方法
```
ChildrenMethods() ([]any, error)
```
例
```
// 嵌套方法
func (m *Main)ChildrenMethods() ([]any, error){
    return []any{
        &Info{},
    },nil
}

type Info struct{
}
func (m *Info) Get(ctx context.Context)(string, error){
    return "jrpc简单示例",nil
}
```
调用
```
req:={"id":3,"jsonrpc":"2.0","method":  "Main.Info.Get","params":null}
res := jsonrpc.Call(context.Background(), []byte(req))
fmt.Println(string(res))
```
返回结果
```
{"jsonrpc":"2.0","id":3,"result":"jrpc简单示例"}
```

http
===

```
http.HandleFunc("/api/endpoint", func(w http.ResponseWriter, r *http.Request) {
    body, err := io.ReadAll(r.Body)
    if err != nil {
        w.WriteHeader(500)
        w.Write([]byte(err.Error()))
    }

    w.Write(jsonrpc.Call(r.Context(), body))
})
http.ListenAndServe(":8000", nil)
```

jrpc_test.go 另有socket实例
