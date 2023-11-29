package goscript

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func parseHttpParams(call FunctionCall) (string, map[string][]string, map[string]string /*map[string]any*/, any) {
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
	body := call.Argument(3).Export()
	//var body map[string]any
	//if _body != nil {
	//	body = _body.(map[string]any)
	//}
	return _url, header, params, body
}

func parseHttpResult(r *Runtime, sc int, header map[string][]string, body any, err error) Value {
	var data map[string]any
	httpRet := map[string]any{
		"statusCode": sc,
		"header":     header,
		"data":       nil,
		"text":       "",
		"error":      "",
	}
	errMsg := ""
	if err != nil {
		errMsg = err.Error()
		httpRet["error"] = errMsg
	} else {
		err = json.Unmarshal(body.([]byte), &data)
		httpRet["text"] = string(body.([]byte))
		if err == nil {
			httpRet["data"] = data
		} else {
			httpRet["error"] = err.Error()
		}
	}
	return r.ToValue(httpRet)
}

func (r *Runtime) builtinHTTP_get(call FunctionCall) Value {
	u, h, p, _ := parseHttpParams(call)
	sc, hd, body, err := privateHttpGet(u, h, p)
	return parseHttpResult(r, sc, hd, body, err)
}

func (r *Runtime) builtinHTTP_post(call FunctionCall) Value {
	u, h, p, b := parseHttpParams(call)
	sc, hd, body, err := privateHttpPost(u, h, p, b)
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
	sc, hd, body, err := privateHttpPostForm(u, h, _p)
	return parseHttpResult(r, sc, hd, body, err)
}

func (r *Runtime) builtinHTTP_put(call FunctionCall) Value {
	u, h, p, b := parseHttpParams(call)
	sc, hd, body, err := privateHttpPut(u, h, p, b)
	return parseHttpResult(r, sc, hd, body, err)
}

func (r *Runtime) builtinHTTP_delete(call FunctionCall) Value {
	u, h, p, _ := parseHttpParams(call)
	sc, hd, body, err := privateHttpDelete(u, h, p)
	return parseHttpResult(r, sc, hd, body, err)
}

func (r *Runtime) builtinHTTP_patch(call FunctionCall) Value {
	u, h, p, b := parseHttpParams(call)
	sc, hd, body, err := privateHttpPatch(u, h, p, b)
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

// migrate from gobase

var httpClient = createHTTPClient()

func createHTTPClient() *http.Client {
	client := &http.Client{}
	// HTTP 客户端配置
	client.Timeout, _ = time.ParseDuration("5s")
	transport := &http.Transport{}
	transport.TLSHandshakeTimeout, _ = time.ParseDuration("10s")
	transport.DisableCompression = true
	transport.MaxIdleConns = 100
	transport.MaxIdleConns = 100
	transport.MaxIdleConnsPerHost = 100
	transport.MaxConnsPerHost = 100
	transport.IdleConnTimeout, _ = time.ParseDuration("90s")
	transport.ResponseHeaderTimeout, _ = time.ParseDuration("15s")
	transport.ExpectContinueTimeout, _ = time.ParseDuration("1s")
	client.Transport = transport
	return client
}

func urlWithParameter(url string, parameterMap map[string]string) string {
	if parameterMap == nil || len(parameterMap) == 0 {
		return url
	}
	url += "?"
	var parameters []string
	for key, value := range parameterMap {
		parameters = append(parameters, key+"="+value)
	}
	return url + strings.Join(parameters, "&")
}

type NetError struct {
	ErrMsg string
}

func (error *NetError) Error() string {
	return error.ErrMsg
}

func doParseResponse(httpResponse *http.Response, err error) (int, http.Header, any, error) {
	if err != nil && httpResponse == nil {
		return -1, nil, nil, &NetError{ErrMsg: "Error sending request, err" + err.Error()}
	} else {
		if httpResponse == nil {
			return -1, nil, nil, nil
		}
		defer func(Body io.ReadCloser) {
			_ = Body.Close()
		}(httpResponse.Body)

		code := httpResponse.StatusCode
		headers := httpResponse.Header
		if code != http.StatusOK {
			body, _ := io.ReadAll(httpResponse.Body)
			return code, headers, nil, &NetError{ErrMsg: "remote error, url: code " + strconv.Itoa(code) + ", message: " + string(body)}
		}
		// We have seen inconsistencies even when we get 200 OK response
		body, err := io.ReadAll(httpResponse.Body)
		if err != nil {
			return code, headers, nil, &NetError{ErrMsg: "Couldn't parse response body, err: " + err.Error()}
		}
		return code, headers, body, nil
	}
}

func callHttp(httpRequest *http.Request) (int, http.Header, any, error) {
	httpResponse, err := httpClient.Do(httpRequest)
	rspCode, rspHead, rspData, err := doParseResponse(httpResponse, err)
	return rspCode, rspHead, rspData, err
}

func privateHttpGet(url string, header http.Header, parameterMap map[string]string) (int, http.Header, any, error) {
	httpRequest, err := http.NewRequest("GET", urlWithParameter(url, parameterMap), nil)
	if err != nil {
		return -1, nil, nil, err
	}
	if header != nil {
		httpRequest.Header = header
	}
	return callHttp(httpRequest)
}

func privateHttpPost(url string, header http.Header, parameterMap map[string]string, body any) (int, http.Header, any, error) {
	bytes, _ := json.Marshal(body)
	payload := strings.NewReader(string(bytes))
	httpRequest, err := http.NewRequest("POST", urlWithParameter(url, parameterMap), payload)
	if err != nil {
		return -1, nil, nil, err
	}
	if header != nil {
		httpRequest.Header = header
	}
	httpRequest.Header.Add("Content-Type", "application/json; charset=utf-8")
	return callHttp(httpRequest)
}

func privateHttpPostForm(url string, header http.Header, parameterMap map[string]any) (int, http.Header, any, error) {
	//先处理将param转body
	var formTempRequest http.Request
	_ = formTempRequest.ParseForm()
	if parameterMap != nil {
		_ = formTempRequest.ParseForm()
		for k, v := range parameterMap {
			formTempRequest.Form.Add(k, fmt.Sprintf("%v", v))
		}
	}
	body := strings.NewReader(formTempRequest.Form.Encode())

	// 开始封装一下真实的请求
	httpReq, err := http.NewRequest("POST", url, body)
	if header != nil {
		httpReq.Header = header
	}
	httpReq.Form = formTempRequest.Form
	httpReq.PostForm = formTempRequest.Form

	if err != nil {
		return 0, nil, nil, err
	}
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := httpClient.Do(httpReq)
	rspCode, rspHead, rspData, err := doParseResponse(resp, err)

	return rspCode, rspHead, rspData, err
}

func privateHttpPut(url string, header http.Header, parameterMap map[string]string, body any) (int, http.Header, any, error) {
	bytes, _ := json.Marshal(body)
	payload := strings.NewReader(string(bytes))
	httpRequest, err := http.NewRequest("PUT", urlWithParameter(url, parameterMap), payload)
	if err != nil {
		return -1, nil, nil, err
	}
	if header != nil {
		httpRequest.Header = header
	}
	httpRequest.Header.Add("Content-Type", "application/json; charset=utf-8")
	return callHttp(httpRequest)
}

func privateHttpDelete(url string, header http.Header, parameterMap map[string]string) (int, http.Header, any, error) {
	httpRequest, err := http.NewRequest("DELETE", urlWithParameter(url, parameterMap), nil)
	if err != nil {
		return -1, nil, nil, err
	}
	if header != nil {
		httpRequest.Header = header
	}
	return callHttp(httpRequest)
}

func privateHttpPatch(url string, header http.Header, parameterMap map[string]string, body any) (int, http.Header, any, error) {
	bytes, _ := json.Marshal(body)
	payload := strings.NewReader(string(bytes))
	httpRequest, err := http.NewRequest("PATCH", urlWithParameter(url, parameterMap), payload)
	if err != nil {
		return -1, nil, nil, err
	}
	if header != nil {
		httpRequest.Header = header
	}
	httpRequest.Header.Add("Content-Type", "application/json; charset=utf-8")
	return callHttp(httpRequest)
}
