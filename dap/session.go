package dap

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/google/go-dap"
	"io"
	"runtime/debug"
	"sync"
)

type Session struct {
	config *Config
	// conn is the accepted client connection.
	conn *connection

	// sendingMu synchronizes writing to conn
	// to ensure that messages do not get interleaved
	sendingMu sync.Mutex
}

func NewSession(conn io.ReadWriteCloser, config *Config) *Session {
	return &Session{
		config: config,
		conn:   &connection{conn, make(chan struct{})},
	}
}

// ServeDAPCodec reads and decodes requests from the client
// until it encounters an error or EOF, when it sends
// a disconnect signal and returns.
func (s *Session) ServeDAPCodec() {
	// Close conn, but not the debugger in case we are in AcceptMuli mode.
	// If not, debugger will be shut down in Stop().
	defer s.conn.Close()
	reader := bufio.NewReader(s.conn)
	for {
		request, err := dap.ReadProtocolMessage(reader)
		// Handle dap.DecodeProtocolMessageFieldError errors gracefully by responding with an ErrorResponse.
		// For example:
		// -- "Request command 'foo' is not supported" means we
		// potentially got some new DAP request that we do not yet have
		// decoding support for, so we can respond with an ErrorResponse.
		//
		// Other errors, such as unmarshalling errors, will log the error and cause the server to trigger
		// a stop.
		if err != nil {
			s.config.log.Debug("DAP error: ", err)
			select {
			case <-s.config.StopTriggered:
			default:
				if err != io.EOF { // EOF means client closed connection
					if decodeErr, ok := err.(*dap.DecodeProtocolMessageFieldError); ok {
						// Send an error response to the users if we were unable to process the message.
						s.sendInternalErrorResponse(decodeErr.Seq, err.Error())
						continue
					}
					s.config.log.Error("DAP error: ", err)
				}
			}
			return
		}
		s.handleRequest(request)

		if _, ok := request.(*dap.DisconnectRequest); ok {
			// disconnect already shut things down and triggered stopping
			return
		}
	}
}

// In case a handler panics, we catch the panic to avoid crashing both
// the server and the target. We send an error response back, but
// in case its a dup and ignored by the client, we also log the error.
func (s *Session) recoverPanic(request dap.Message) {
	if ierr := recover(); ierr != nil {
		s.config.log.Errorf("recovered panic: %s\n%s\n", ierr, debug.Stack())
		s.sendInternalErrorResponse(request.GetSeq(), fmt.Sprintf("%v", ierr))
	}
}

func (s *Session) handleRequest(request dap.Message) {
	defer s.recoverPanic(request)

	jsonmsg, _ := json.Marshal(request)
	s.config.log.Debug("[<- from client]", string(jsonmsg))

	if _, ok := request.(dap.RequestMessage); !ok {
		s.sendInternalErrorResponse(request.GetSeq(), fmt.Sprintf("Unable to process non-request %#v\n", request))
		return
	}
	if s.checkNoDebug(request) {
		return
	}
	switch request := request.(type) {
	case *dap.InitializeRequest: // Required
		s.onInitializeRequest(request)
	case *dap.LaunchRequest: // Required
		s.onLaunchRequest(request)
	case *dap.AttachRequest: // Required
		s.onAttachRequest(request)
	case *dap.DisconnectRequest: // Required
		s.onDisconnectRequest(request)
	case *dap.PauseRequest: // Required
		s.onPauseRequest(request)
	//case *dap.TerminateRequest: // Optional (capability ‘supportsTerminateRequest‘)
	//case *dap.RestartRequest: // Optional (capability ‘supportsRestartRequest’)

	//--- 【Asynchronous requests】 ---
	//case *dap.ConfigurationDoneRequest: // Optional (capability ‘supportsConfigurationDoneRequest’)
	case *dap.ContinueRequest: // Required
		s.onContinueRequest(request)
	case *dap.NextRequest: // Required
		s.onNextRequest(request)
	case *dap.StepInRequest: // Required
		s.onStepInRequest(request)
	case *dap.StepOutRequest: // Required
		s.onStepOutRequest(request)
	//case *dap.StepBackRequest: // Optional (capability ‘supportsStepBack’)
	//case *dap.ReverseContinueRequest: // Optional (capability ‘supportsStepBack’)

	//--- 【Synchronous requests】 ---
	case *dap.SetBreakpointsRequest: // Required
		s.onSetBreakpointsRequest(request)
	//case *dap.SetFunctionBreakpointsRequest: // Optional (capability ‘supportsFunctionBreakpoints’)
	//case *dap.SetInstructionBreakpointsRequest: // Optional (capability 'supportsInstructionBreakpoints')
	//case *dap.SetExceptionBreakpointsRequest: // Optional (capability ‘exceptionBreakpointFilters’)
	case *dap.ThreadsRequest: // Required
		s.onThreadsRequest(request)
	case *dap.StackTraceRequest: // Required
		s.onStackTraceRequest(request)
	case *dap.ScopesRequest: // Required
		s.onScopesRequest(request)
	case *dap.VariablesRequest: // Required
		s.onVariablesRequest(request)
	case *dap.EvaluateRequest: // Required
		s.onEvaluateRequest(request)
	//case *dap.SetVariableRequest: // Optional (capability ‘supportsSetVariable’)
	//case *dap.ExceptionInfoRequest: // Optional (capability ‘supportsExceptionInfoRequest’)
	//case *dap.DisassembleRequest: // Optional (capability ‘supportsDisassembleRequest’)
	default:
		s.sendInternalErrorResponse(request.GetSeq(), fmt.Sprintf("Unable to process %#v\n", request))
	}
}

// sendInternalErrorResponse sends an "internal error" response back to the client.
// We only take a seq here because we don't want to make assumptions about the
// kind of message received by the server that this error is a reply to.
func (s *Session) sendInternalErrorResponse(seq int, details string) {
	er := &dap.ErrorResponse{}
	er.Type = "response"
	er.RequestSeq = seq
	er.Success = false
	er.Message = "Internal Error"
	er.Body.Error.Id = InternalError
	er.Body.Error.Format = fmt.Sprintf("%s: %s", er.Message, details)
	s.config.log.Debug(er.Body.Error.Format)
	s.send(er)
}
func (s *Session) sendUnsupportedErrorResponse(request dap.Request) {
	s.sendErrorResponse(request, UnsupportedCommand, "Unsupported command",
		fmt.Sprintf("cannot process %q request", request.Command))
}

// sendErrorResponse sends an error response with showUser disabled (default).
func (s *Session) sendErrorResponse(request dap.Request, id int, summary, details string) {
	s.sendErrorResponseWithOpts(request, id, summary, details, false /*showUser*/)
}

// sendErrorResponseWithOpts offers configuration options.
//
//	showUser - if true, the error will be shown to the user (e.g. via a visible pop-up)
func (s *Session) sendErrorResponseWithOpts(request dap.Request, id int, summary, details string, showUser bool) {
	er := &dap.ErrorResponse{}
	er.Type = "response"
	er.Command = request.Command
	er.RequestSeq = request.Seq
	er.Success = false
	er.Message = summary
	er.Body.Error.Id = id
	er.Body.Error.Format = fmt.Sprintf("%s: %s", summary, details)
	er.Body.Error.ShowUser = showUser
	s.config.log.Debug(er.Body.Error.Format)
	s.send(er)
}

func (s *Session) send(message dap.Message) {
	jsonmsg, _ := json.Marshal(message)
	s.config.log.Debug("[-> to client]", string(jsonmsg))

	// TODO(polina): consider using a channel for all the sends and to have a dedicated
	// goroutine that reads from that channel and sends over the connection.
	// This will avoid blocking on slow network sends.
	s.sendingMu.Lock()
	defer s.sendingMu.Unlock()
	err := dap.WriteProtocolMessage(s.conn, message)
	if err != nil {
		s.config.log.Debug(err)
	}
}

func (s Session) Close() {

}

func (s *Session) checkNoDebug(request dap.Message) bool {
	//if debug return false

	switch request := request.(type) {
	case *dap.DisconnectRequest:
		//todo 断开连接
		s.onDisconnectRequest(request)
	case *dap.RestartRequest:
		s.sendUnsupportedErrorResponse(request.Request)
	default:
		r := request.(dap.RequestMessage).GetRequest()
		s.sendErrorResponse(*r, NoDebugIsRunning, "noDebug mode", fmt.Sprintf("unable to process '%s' request", r.Command))
	}
	return true
}

type connection struct {
	io.ReadWriteCloser
	closed chan struct{}
}

func (c *connection) Close() error {
	select {
	case <-c.closed:
	default:
		close(c.closed)
	}
	return c.ReadWriteCloser.Close()
}

func (c *connection) isClosed() bool {
	select {
	case <-c.closed:
		return true
	default:
		return false
	}
}
