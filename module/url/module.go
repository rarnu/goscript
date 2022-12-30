package url

import (
	"math"
	"net/url"
	"reflect"
	"strconv"
	"strings"

	"golang.org/x/net/idna"

	"github.com/rarnu/goscript"
	"github.com/rarnu/goscript/module/require"
)

const ModuleName = "node:url"

var (
	reflectTypeURL = reflect.TypeOf((*url.URL)(nil))
	reflectTypeInt = reflect.TypeOf(0)
)

func isDefaultURLPort(protocol string, port int) bool {
	switch port {
	case 21:
		if protocol == "ftp" {
			return true
		}
	case 80:
		if protocol == "http" || protocol == "ws" {
			return true
		}
	case 443:
		if protocol == "https" || protocol == "wss" {
			return true
		}
	}
	return false
}

func isSpecialProtocol(protocol string) bool {
	switch protocol {
	case "ftp", "file", "http", "https", "ws", "wss":
		return true
	}
	return false
}

func clearURLPort(u *url.URL) {
	u.Host = u.Hostname()
}

func valueToURLPort(v goscript.Value) (portNum int, empty bool) {
	portNum = -1
	if et := v.ExportType(); et == reflectTypeInt {
		if num := v.ToInteger(); num >= 0 && num <= math.MaxUint16 {
			portNum = int(num)
		}
	} else {
		s := v.String()
		if s == "" {
			return 0, true
		}
		for i := 0; i < len(s); i++ {
			if c := s[i]; c >= '0' && c <= '9' {
				if portNum == -1 {
					portNum = 0
				}
				portNum = portNum*10 + int(c-'0')
				if portNum > math.MaxUint16 {
					portNum = -1
					break
				}
			} else {
				break
			}
		}
	}
	return
}

func setURLPort(u *url.URL, v goscript.Value) {
	if u.Scheme == "file" {
		return
	}
	portNum, empty := valueToURLPort(v)
	if empty {
		clearURLPort(u)
		return
	}
	if portNum == -1 {
		return
	}
	if isDefaultURLPort(u.Scheme, portNum) {
		clearURLPort(u)
	} else {
		u.Host = u.Hostname() + ":" + strconv.Itoa(portNum)
	}
}

func toURL(r *goscript.Runtime, v goscript.Value) *url.URL {
	if v.ExportType() == reflectTypeURL {
		if u := v.Export().(*url.URL); u != nil {
			return u
		}
	}
	panic(r.NewTypeError("Expected URL"))
}

func defineURLAccessorProp(r *goscript.Runtime, p *goscript.Object, name string, getter func(*url.URL) interface{}, setter func(*url.URL, goscript.Value)) {
	var getterVal, setterVal goscript.Value
	if getter != nil {
		getterVal = r.ToValue(func(call goscript.FunctionCall) goscript.Value {
			return r.ToValue(getter(toURL(r, call.This)))
		})
	}
	if setter != nil {
		setterVal = r.ToValue(func(call goscript.FunctionCall) goscript.Value {
			setter(toURL(r, call.This), call.Argument(0))
			return goscript.Undefined()
		})
	}
	p.DefineAccessorProperty(name, getterVal, setterVal, goscript.FLAG_FALSE, goscript.FLAG_TRUE)
}

func createURLPrototype(r *goscript.Runtime) *goscript.Object {
	p := r.NewObject()

	// host
	defineURLAccessorProp(r, p, "host", func(u *url.URL) interface{} {
		return u.Host
	}, func(u *url.URL, arg goscript.Value) {
		host := arg.String()
		if _, err := url.ParseRequestURI(u.Scheme + "://" + host); err == nil {
			u.Host = host
			fixURL(r, u)
		}
	})

	// hash
	defineURLAccessorProp(r, p, "hash", func(u *url.URL) interface{} {
		if u.Fragment != "" {
			return "#" + u.EscapedFragment()
		}
		return ""
	}, func(u *url.URL, arg goscript.Value) {
		h := arg.String()
		if len(h) > 0 && h[0] == '#' {
			h = h[1:]
		}
		u.Fragment = h
	})

	// hostname
	defineURLAccessorProp(r, p, "hostname", func(u *url.URL) interface{} {
		return strings.Split(u.Host, ":")[0]
	}, func(u *url.URL, arg goscript.Value) {
		h := arg.String()
		if strings.IndexByte(h, ':') >= 0 {
			return
		}
		if _, err := url.ParseRequestURI(u.Scheme + "://" + h); err == nil {
			if port := u.Port(); port != "" {
				u.Host = h + ":" + port
			} else {
				u.Host = h
			}
			fixURL(r, u)
		}
	})

	// href
	defineURLAccessorProp(r, p, "href", func(u *url.URL) interface{} {
		return u.String()
	}, func(u *url.URL, arg goscript.Value) {
		url := parseURL(r, arg.String(), true)
		*u = *url
	})

	// pathname
	defineURLAccessorProp(r, p, "pathname", func(u *url.URL) interface{} {
		return u.EscapedPath()
	}, func(u *url.URL, arg goscript.Value) {
		p := arg.String()
		if _, err := url.Parse(p); err == nil {
			switch u.Scheme {
			case "https", "http", "ftp", "ws", "wss":
				if !strings.HasPrefix(p, "/") {
					p = "/" + p
				}
			}
			u.Path = p
		}
	})

	// origin
	defineURLAccessorProp(r, p, "origin", func(u *url.URL) interface{} {
		return u.Scheme + "://" + u.Hostname()
	}, nil)

	// password
	defineURLAccessorProp(r, p, "password", func(u *url.URL) interface{} {
		p, _ := u.User.Password()
		return p
	}, func(u *url.URL, arg goscript.Value) {
		user := u.User
		u.User = url.UserPassword(user.Username(), arg.String())
	})

	// username
	defineURLAccessorProp(r, p, "username", func(u *url.URL) interface{} {
		return u.User.Username()
	}, func(u *url.URL, arg goscript.Value) {
		p, has := u.User.Password()
		if !has {
			u.User = url.User(arg.String())
		} else {
			u.User = url.UserPassword(arg.String(), p)
		}
	})

	// port
	defineURLAccessorProp(r, p, "port", func(u *url.URL) interface{} {
		return u.Port()
	}, func(u *url.URL, arg goscript.Value) {
		setURLPort(u, arg)
	})

	// protocol
	defineURLAccessorProp(r, p, "protocol", func(u *url.URL) interface{} {
		return u.Scheme + ":"
	}, func(u *url.URL, arg goscript.Value) {
		s := arg.String()
		pos := strings.IndexByte(s, ':')
		if pos >= 0 {
			s = s[:pos]
		}
		s = strings.ToLower(s)
		if isSpecialProtocol(u.Scheme) == isSpecialProtocol(s) {
			if _, err := url.ParseRequestURI(s + "://" + u.Host); err == nil {
				u.Scheme = s
			}
		}
	})

	// Search
	defineURLAccessorProp(r, p, "search", func(u *url.URL) interface{} {
		if u.RawQuery != "" {
			return "?" + u.RawQuery
		}
		return ""
	}, func(u *url.URL, arg goscript.Value) {
		u.RawQuery = arg.String()
		fixRawQuery(u)
	})

	p.Set("toString", r.ToValue(func(call goscript.FunctionCall) goscript.Value {
		return r.ToValue(toURL(r, call.This).String())
	}))

	p.Set("toJSON", r.ToValue(func(call goscript.FunctionCall) goscript.Value {
		return r.ToValue(toURL(r, call.This).String())
	}))

	return p
}

const (
	URLNotAbsolute  = "URL is not absolute"
	InvalidURL      = "Invalid URL"
	InvalidBaseURL  = "Invalid base URL"
	InvalidHostname = "Invalid hostname"
)

func newInvalidURLError(r *goscript.Runtime, msg, input string) *goscript.Object {
	// when node's error module is added this should return a NodeError
	o := r.NewTypeError(msg)
	o.Set("input", r.ToValue(input))
	return o
}

func fixRawQuery(u *url.URL) {
	if u.RawQuery != "" {
		var u1 url.URL
		u1.Fragment = u.RawQuery
		u.RawQuery = u1.EscapedFragment()
	}
}

func fixURL(r *goscript.Runtime, u *url.URL) {
	switch u.Scheme {
	case "https", "http", "ftp", "wss", "ws":
		if u.Path == "" {
			u.Path = "/"
		}
		hostname := u.Hostname()
		lh := strings.ToLower(hostname)
		ch, err := idna.Punycode.ToASCII(lh)
		if err != nil {
			panic(newInvalidURLError(r, InvalidHostname, lh))
		}
		if ch != hostname {
			if port := u.Port(); port != "" {
				u.Host = ch + ":" + port
			} else {
				u.Host = ch
			}
		}
		fixRawQuery(u)
	}
}

func parseURL(r *goscript.Runtime, s string, isBase bool) *url.URL {
	u, err := url.Parse(s)
	if err != nil {
		if isBase {
			panic(newInvalidURLError(r, InvalidBaseURL, s))
		} else {
			panic(newInvalidURLError(r, InvalidURL, s))
		}
	}
	if isBase && !u.IsAbs() {
		panic(newInvalidURLError(r, URLNotAbsolute, s))
	}
	if portStr := u.Port(); portStr != "" {
		if port, err := strconv.Atoi(portStr); err != nil || isDefaultURLPort(u.Scheme, port) {
			clearURLPort(u)
		}
	}
	fixURL(r, u)
	return u
}

func createURLConstructor(r *goscript.Runtime) goscript.Value {
	f := r.ToValue(func(call goscript.ConstructorCall) *goscript.Object {
		var u *url.URL
		if baseArg := call.Argument(1); !goscript.IsUndefined(baseArg) {
			base := parseURL(r, baseArg.String(), true)
			ref := parseURL(r, call.Arguments[0].String(), false)
			u = base.ResolveReference(ref)
		} else {
			u = parseURL(r, call.Argument(0).String(), true)
		}
		res := r.ToValue(u).(*goscript.Object)
		res.SetPrototype(call.This.Prototype())
		return res
	}).(*goscript.Object)

	f.Set("prototype", createURLPrototype(r))
	return f
}

func Require(runtime *goscript.Runtime, module *goscript.Object) {
	exports := module.Get("exports").(*goscript.Object)
	exports.Set("URL", createURLConstructor(runtime))
}

func Enable(runtime *goscript.Runtime) {
	runtime.Set("URL", require.Require(runtime, ModuleName).ToObject(runtime).Get("URL"))
}

func init() {
	require.RegisterNativeModule(ModuleName, Require)
}
