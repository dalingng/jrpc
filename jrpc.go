package jrpc

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
)

type Error interface {
	Error() string
	GetCode() int
	GetData() any
}

type Request struct {
	Jsonrpc json.RawMessage `json:"jsonrpc"`
	Id      json.RawMessage `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type ResponseError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

func (re *ResponseError) Error() string {
	return re.Message
}

func (re *ResponseError) GetCode() int {
	return re.Code
}

func (re *ResponseError) GetData() any {
	return re.Data
}

func NewError(code int, message string, data any) *ResponseError {
	return &ResponseError{
		Code:    code,
		Message: message,
		Data:    data,
	}
}

type Response struct {
	Jsonrpc json.RawMessage `json:"jsonrpc"`
	Id      json.RawMessage `json:"id"`
	Result  any             `json:"result"`
	Error   *ResponseError  `json:"error,omitempty"`
}

type RPCMethod struct {
	rval   reflect.Value
	method reflect.Method
	name   string
}
type HasChildren interface {
	ChildrenMethods() ([]any, error)
}
type HasAliasName interface {
	AliasMethodName() string
}

type JSONRPC map[string]*RPCMethod

func (rpc JSONRPC) Register(m any, prefix ...string) error {
	rval := reflect.ValueOf(m)
	rtype := reflect.TypeOf(m)

	// 提取名称
	var sName string
	switch v := m.(type) {
	case HasAliasName:
		sName = v.AliasMethodName()
	default:
		if rtype.Kind() == reflect.Pointer {
			sName = rtype.Elem().Name()
		} else {
			sName = rtype.Name()
		}
	}

	// 提取方法
	for i := 0; i < rtype.NumMethod(); i++ {
		mName := rtype.Method(i).Name
		m, _ := rtype.MethodByName(mName)

		methodName := strings.Join(append(prefix, sName, mName), ".")
		// 把方法记录到method
		method := &RPCMethod{
			rval:   rval,
			method: m,
			name:   methodName,
		}

		rpc[method.name] = method
	}
	switch v := m.(type) {
	case HasChildren:
		mm, err := v.ChildrenMethods()
		if err != nil {
			return err
		}
		return rpc.RegisterMultiple(mm, append(prefix, sName)...)
	default:
	}
	return nil
}

func (rpc JSONRPC) RegisterMultiple(ms []any, prefix ...string) error {
	for _, v := range ms {
		err := rpc.Register(v, prefix...)
		if err != nil {
			return err
		}
	}
	return nil
}

func (rpc JSONRPC) GetMethod(name string) *RPCMethod {
	return rpc[name]
}
func (rpc JSONRPC) GetMethods() []string {
	methods := []string{}
	for _, v := range rpc {
		methods = append(methods, v.name)
	}
	return methods
}

func NewResponseError(req *Request, code int, message string, data any) *Response {
	return &Response{
		Id:      req.Id,
		Jsonrpc: req.Jsonrpc,
		Error: &ResponseError{
			Code:    code,
			Message: message,
		},
	}
}

func (rpc JSONRPC) CallJson(ctx context.Context, jsonData []byte) ([]byte, error) {
	if len(jsonData) == 0 {
		return []byte(""), errors.New("请求数据不能为空")
	}
	if jsonData[0] == '[' {
		reqs := []*Request{}
		err := json.Unmarshal(jsonData, &reqs)
		if err != nil {
			return []byte(err.Error()), err
		}
		resps := []*Response{}
		for _, req := range reqs {
			r := rpc.Call(ctx, req)
			resps = append(resps, r)
		}
		b, err := json.Marshal(resps)
		return b, err
	}
	req := &Request{}
	err := json.Unmarshal(jsonData, req)
	if err != nil {
		return []byte(err.Error()), errors.ErrUnsupported
	}
	resp := rpc.Call(ctx, req)
	b, _ := json.Marshal(resp)
	return b, nil
}

func (rpc JSONRPC) Call(ctx context.Context, req *Request) *Response {
	methodName := req.Method
	rpcMethod := rpc.GetMethod(methodName)

	if methodName == "Methods" {
		return GetMethods(rpc, req)
	}

	if rpcMethod == nil {
		return NewResponseError(req, -32601, "程序错误:"+methodName+"方法不存在", nil)
	}

	if rpcMethod.method.Type.NumIn() > 3 {
		return NewResponseError(req, -32601, "程序错误:方法参数数量不对", nil)
	}
	if rpcMethod.method.Type.NumOut() != 2 {
		return NewResponseError(req, -32601, "程序错误:方法返回数量不对", nil)
	}

	params := make([]reflect.Value, rpcMethod.method.Type.NumIn())

	// 结构体自己
	params[0] = rpcMethod.rval

	// context
	params[1] = reflect.ValueOf(ctx)

	if rpcMethod.method.Type.NumIn() == 3 {
		// params
		argType := rpcMethod.method.Type.In(2)
		var argv reflect.Value
		var elem any
		if argType.Kind() != reflect.Pointer {
			argv = reflect.New(argType).Elem()
			elem = argv.Addr().Interface()
		} else {
			argv = reflect.New(argType.Elem())
			elem = argv.Interface()
		}

		err := json.Unmarshal(req.Params, elem)
		if err != nil {
			return NewResponseError(req, -32601, err.Error(), nil)
		}
		params[2] = argv
	}

	r := rpcMethod.method.Func.Call(params)

	// 返回的params
	res := r[0].Interface()

	// 处理返回的错误
	re := r[1].Interface()

	var rErr *ResponseError
	switch i := re.(type) {
	case ResponseError:
		rErr = &i
	case *ResponseError:
		rErr = i
	case Error:
		rErr = &ResponseError{
			Code:    i.GetCode(),
			Message: i.Error(),
			Data:    i.GetData(),
		}
	case error:
		rErr = &ResponseError{
			Code:    500,
			Message: i.Error(),
			Data:    nil,
		}
	}

	resp := &Response{
		Id:      req.Id,
		Jsonrpc: req.Jsonrpc,
		Result:  res,
		Error:   rErr,
	}
	// b, _ := json.Marshal(resp)
	return resp
}
