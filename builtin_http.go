package goscript

import (
	"encoding/json"
	"fmt"
	h0 "github.com/isyscore/isc-gobase/http"
)

func parseHttpParams(call FunctionCall) (string, map[string][]string, map[string]string, map[string]any) {
	_url := call.Argument(0).toString().String()
	_header := call.Argument(1).Export()
	_params := call.Argument(2).Export()
	header := map[string][]string{}
	if _header != nil {
		for k, v := range _header.(map[string]any) {
			header[k] = []string{fmt.Sprintf("%v", v)}
		}
	}
	params := map[string]string{}
	if _params != nil {
		for k, v := range _params.(map[string]any) {
			params[k] = fmt.Sprintf("%v", v)
		}
	}
	_body := call.Argument(3).Export()
	var body map[string]any
	if _body != nil {
		body = _body.(map[string]any)
	}
	return _url, header, params, body
}

func parseHttpResult(r *Runtime, sc int, header map[string][]string, body any, err error) Value {
	var data map[string]any
	errMsg := ""
	if err != nil {
		errMsg = err.Error()
	} else {
		_ = json.Unmarshal(body.([]byte), &data)
	}
	return r.ToValue(map[string]any{
		"statusCode": sc,
		"header":     header,
		"data":       data,
		"text":       string(body.([]byte)),
		"error":      errMsg,
	})
}

func (r *Runtime) builtinHTTP_get(call FunctionCall) Value {
	u, h, p, _ := parseHttpParams(call)
	sc, hd, body, err := h0.Get(u, h, p)
	return parseHttpResult(r, sc, hd, body, err)
}

func (r *Runtime) builtinHTTP_post(call FunctionCall) Value {
	u, h, p, b := parseHttpParams(call)
	sc, hd, body, err := h0.Post(u, h, p, b)
	return parseHttpResult(r, sc, hd, body, err)
}

func (r *Runtime) builtinHTTP_postForm(call FunctionCall) Value {
	u, h, p, _ := parseHttpParams(call)
	_p := map[string]any{}
	if p != nil {
		for k, v := range p {
			_p[k] = v
		}
	}
	sc, hd, body, err := h0.PostForm(u, h, _p)
	return parseHttpResult(r, sc, hd, body, err)
}

func (r *Runtime) builtinHTTP_put(call FunctionCall) Value {
	u, h, p, b := parseHttpParams(call)
	sc, hd, body, err := h0.Put(u, h, p, b)
	return parseHttpResult(r, sc, hd, body, err)
}

func (r *Runtime) builtinHTTP_delete(call FunctionCall) Value {
	u, h, p, _ := parseHttpParams(call)
	sc, hd, body, err := h0.Delete(u, h, p)
	return parseHttpResult(r, sc, hd, body, err)
}

func (r *Runtime) builtinHTTP_patch(call FunctionCall) Value {
	u, h, p, b := parseHttpParams(call)
	sc, hd, body, err := h0.Patch(u, h, p, b)
	return parseHttpResult(r, sc, hd, body, err)
}

func (r *Runtime) initHttp() {
	HTTP := r.newBaseObject(r.global.ObjectPrototype, "HTTP")
	HTTP._putProp("get", r.newNativeFunc(r.builtinHTTP_get, nil, "get", nil, 3), true, false, true)
	HTTP._putProp("post", r.newNativeFunc(r.builtinHTTP_post, nil, "post", nil, 4), true, false, true)
	HTTP._putProp("postForm", r.newNativeFunc(r.builtinHTTP_postForm, nil, "postForm", nil, 4), true, false, true)
	HTTP._putProp("put", r.newNativeFunc(r.builtinHTTP_put, nil, "put", nil, 4), true, false, true)
	HTTP._putProp("delete", r.newNativeFunc(r.builtinHTTP_delete, nil, "delete", nil, 3), true, false, true)
	HTTP._putProp("patch", r.newNativeFunc(r.builtinHTTP_patch, nil, "patch", nil, 4), true, false, true)

	r.addToGlobal("HTTP", HTTP.val)
}
