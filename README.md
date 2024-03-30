关于jrpc
===
jrpc是一个非常简单JSONRPC实现

程序不考虑性能只考虑简单实现，实测性能还行

使用详情请看 jrpc_test.go

使用方法([]byte请求)
---

```golang
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

```golang
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
```json
[{"jsonrpc":"2.0","id":1,"result":"","error":{"code":10,"message":"请填写用户名"}},{"jsonrpc":"2.0","id":2,"result":"你好:大宁"}]
{"jsonrpc":"2.0","id":2,"result":"你好:单个请求"}
```

入参和出参支持任何类型的数据
```golang
func (m *Main) HandleStr(ctx context.Context, req string) (string, error) {
	return req, nil
}

func (m *Main) HandleStrArr(ctx context.Context, req []string) ([]string, error) {
	return req, nil
}
// 或反回any
func (m *Main) HandleStrArr(ctx context.Context, req int) (any, error) {
	return req, nil
}
```
错误处理
```golang
func (m *Main) HandleErr(ctx context.Context, req *int) (int, error) {
    if req==nil{
        return 0,errors.New("参数不能为空")
    }
	return req, nil
}
// {"jsonrpc":2.0,"id":1,"result":0,"error":{"code":500,"message":"参数不能为空"}}

// 也可以用自定义的jsonrpc错误格式
func (m *Main) HandleJsonErr(ctx context.Context, req struct{
    Num *int
}) (int, error) {
    if req.Num==nil{
        return 0,jrpc.NewError(412, "Num参数不能为空", map[string]string{"Num":"Num不能为空"})
    }
	return req.Num, nil
}
// {"jsonrpc":2.0,"id":1,"result":0,"error":{"code":412,"message":"Num参数不能为空","data":{"Num":"Num不能为空"}}}
```

嵌套
===
需要实现方法
```golang
ChildrenMethods() ([]any, error)
```
例
```golang
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
```golang
req:={"id":3,"jsonrpc":"2.0","method":  "Main.Info.Get","params":null}
res := jsonrpc.Call(context.Background(), []byte(req))
fmt.Println(string(res))
```
返回结果
```json
{"jsonrpc":"2.0","id":3,"result":"jrpc简单示例"}
```

http
===

```golang
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
