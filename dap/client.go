package dap

import (
	"bufio"
	"encoding/json"
	"errors"
	"github.com/google/go-dap"
	"net"
	"sync/atomic"
	"time"
)

func NewClient(address string) (*Client, error) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return nil, err
	}
	c := &Client{
		conn:     conn,
		reader:   bufio.NewReader(conn),
		readChan: make(chan ReadBody),
		seq:      atomic.Int32{},
	}
	go c.Read()
	return c, nil
}

type Client struct {
	conn     net.Conn
	reader   *bufio.Reader
	seq      atomic.Int32
	readChan chan ReadBody
	MsgChan  chan string
}

type ReadBody struct {
	err   error
	bytes []byte
}

func (c *Client) Read() {
	for {
		_ = c.conn.SetReadDeadline(time.Now().Add(5 * time.Second))
		b, err := dap.ReadBaseMessage(c.reader)
		if err != nil {
			c.readChan <- ReadBody{
				err:   err,
				bytes: nil,
			}
		} else {
			var msg dap.ProtocolMessage
			if err := json.Unmarshal(b, &msg); err == nil {
				if msg.Type == "response" {
					c.readChan <- ReadBody{
						err:   nil,
						bytes: b,
					}
				} else if msg.Type == "event" {
					// event := "【event】: " + string(b)
					if c.MsgChan != nil {
						c.MsgChan <- string(b)
					}
				} else {
					if c.MsgChan != nil {
						c.MsgChan <- string(b)
					}
				}
			}
		}
	}
}

func (c *Client) Close() error {
	if c.conn == nil {
		return nil
	}
	return c.conn.Close()
}

func (c *Client) sendAndReceive(req dap.RequestMessage, res any) error {
	request := req.GetRequest()
	request.Type = "request"
	request.Seq = int(c.seq.Add(1))

	b, err := json.Marshal(req)
	if err != nil {
		return err
	}

	_ = c.conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	if err = dap.WriteBaseMessage(c.conn, b); err != nil {
		return err
	}

	select {
	case <-time.NewTimer(5 * time.Second).C:
		return errors.New(" 5s read time out")
	case readBody := <-c.readChan:
		if readBody.err != nil {
			return err
		}
		if err = json.Unmarshal(readBody.bytes, res); err != nil {
			return err
		}
	}

	return nil
}
func (c *Client) Initialize(req *dap.InitializeRequest) (*dap.InitializeResponse, error) {
	req.Command = "initialize"
	var res = &dap.InitializeResponse{}
	if err := c.sendAndReceive(req, res); err != nil {
		return nil, err
	}
	return res, nil
}

func (c *Client) Launch(req *dap.LaunchRequest) (*dap.LaunchResponse, error) {
	req.Command = "launch"
	var res = &dap.LaunchResponse{}
	if err := c.sendAndReceive(req, res); err != nil {
		return nil, err
	}
	return res, nil
}

func (c *Client) OnContinue(req *dap.ContinueRequest) (*dap.ContinueResponse, error) {
	req.Command = "continue"
	var res = &dap.ContinueResponse{}
	if err := c.sendAndReceive(req, res); err != nil {
		return nil, err
	}
	return res, nil
}

func (c *Client) OnNext(req *dap.NextRequest) (*dap.NextResponse, error) {
	req.Command = "next"
	var res = &dap.NextResponse{}
	if err := c.sendAndReceive(req, res); err != nil {
		return nil, err
	}
	return res, nil
}
func (c *Client) onStepIn(req *dap.StepInRequest) (*dap.StepOutResponse, error) {
	req.Command = "stepIn"
	var res = &dap.StepOutResponse{}
	if err := c.sendAndReceive(req, res); err != nil {
		return nil, err
	}
	return res, nil
}

func (c *Client) onStepOut(req *dap.StepOutRequest) (*dap.StepOutResponse, error) {
	req.Command = "stepOut"
	var res = &dap.StepOutResponse{}
	if err := c.sendAndReceive(req, res); err != nil {
		return nil, err
	}
	return res, nil
}

func (c *Client) SetBreakpoints(req *dap.SetBreakpointsRequest) (*dap.SetBreakpointsResponse, error) {
	req.Command = "setBreakpoints"
	var res = &dap.SetBreakpointsResponse{}
	if err := c.sendAndReceive(req, res); err != nil {
		return nil, err
	}
	return res, nil
}

func (c *Client) OnVariables(req *dap.VariablesRequest) (*dap.VariablesResponse, error) {
	req.Command = "variables"
	var res = &dap.VariablesResponse{}
	if err := c.sendAndReceive(req, res); err != nil {
		return nil, err
	}
	return res, nil
}

func (c *Client) Evaluate(req *dap.EvaluateRequest) (*dap.EvaluateResponse, error) {
	req.Command = "evaluate"
	var res = &dap.EvaluateResponse{}
	if err := c.sendAndReceive(req, res); err != nil {
		return nil, err
	}
	return res, nil
}

func (c *Client) OnDisconnectRequest(req *dap.DisconnectRequest) (*dap.DisconnectResponse, error) {
	req.Command = "disconnect"
	var res = &dap.DisconnectResponse{}
	if err := c.sendAndReceive(req, res); err != nil {
		return nil, err
	}
	return res, nil
}
