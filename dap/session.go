package dap

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/google/go-dap"
	"github.com/rarnu/goscript"
	"io"
	"os"
	"path/filepath"
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

	// mu synchronizes access to objects set on start-up (from run goroutine)
	// and stopped on teardown (from main goroutine)
	mu sync.Mutex

	r   *goscript.Runtime
	prg *goscript.Program
}

func debugRun() error {
	s := NewSession(nil, &Config{})

	// receive command from terminal
	commandCh := make(chan string)

	for command := range commandCh {
		request, err := dap.DecodeProtocolMessage([]byte(command))
		if err != nil {
			return err
		}
		// dap.LaunchRequest => go s.r.compile("")
		// dap.AttachRequest => new debugger

		// send response to terminal
		//if err = s.handleRequest(request); err != nil {
		//	// stop script
		//	return err
		//}
	}
	// stop script

	return nil
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
		// init Runtime
		// response support capability with dap
		s.onInitializeRequest(request)
	case *dap.LaunchRequest: // Required
		// compile script
		// if no debug start ,else init debugger
		s.onLaunchRequest(request)
	case *dap.AttachRequest: // Required
		// record ProcessID
		// init debugger
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

	// It is automatically sent as part of the threads > stacktrace > scopes > variables
	// "waterfall" to highlight the topmost frame at stops, after an evaluate request
	// for the selected scope or when a user selects different scopes in the UI.
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

// sendShowUserErrorResponse sends an error response with showUser enabled.
func (s *Session) sendShowUserErrorResponse(request dap.Request, id int, summary, details string) {
	s.sendErrorResponseWithOpts(request, id, summary, details, true /*showUser*/)
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
	if s.conn != nil && !s.conn.isClosed() {
		s.conn.Close()
	}
}

func (s *Session) checkNoDebug(request dap.Message) bool {
	//if debug return false

	switch request := request.(type) {
	case *dap.DisconnectRequest:
		s.onDisconnectRequest(request)
	case *dap.RestartRequest:
		s.sendUnsupportedErrorResponse(request.Request)
	default:
		r := request.(dap.RequestMessage).GetRequest()
		s.sendErrorResponse(*r, NoDebugIsRunning, "noDebug mode", fmt.Sprintf("unable to process '%s' request", r.Command))
	}
	return true
}

func (s *Session) onInitializeRequest(request *dap.InitializeRequest) {
	if request.Arguments.PathFormat != "path" {
		s.sendErrorResponse(request.Request, FailedToInitialize, "Failed to initialize",
			fmt.Sprintf("Unsupported 'pathFormat' value '%s'.", request.Arguments.PathFormat))
		return
	}
	if !request.Arguments.LinesStartAt1 {
		s.sendErrorResponse(request.Request, FailedToInitialize, "Failed to initialize",
			"Only 1-based line numbers are supported.")
		return
	}
	if !request.Arguments.ColumnsStartAt1 {
		s.sendErrorResponse(request.Request, FailedToInitialize, "Failed to initialize",
			"Only 1-based column numbers are supported.")
		return
	}

	s.r = goscript.New()

	response := &dap.InitializeResponse{Response: *newResponse(request.Request)}
	response.Body.SupportsConfigurationDoneRequest = true
	response.Body.SupportsConditionalBreakpoints = true
	response.Body.SupportsDelayedStackTraceLoading = true
	response.Body.SupportsFunctionBreakpoints = true
	response.Body.SupportsInstructionBreakpoints = true
	response.Body.SupportsExceptionInfoRequest = true
	response.Body.SupportsSetVariable = true
	response.Body.SupportsEvaluateForHovers = true
	response.Body.SupportsClipboardContext = true
	response.Body.SupportsSteppingGranularity = true
	response.Body.SupportsLogPoints = true
	response.Body.SupportsDisassembleRequest = true
	// To be enabled by CapabilitiesEvent based on launch configuration
	response.Body.SupportsStepBack = false
	response.Body.SupportTerminateDebuggee = false
	// TODO(polina): support these requests in addition to vscode-go feature parity
	response.Body.SupportsTerminateRequest = false
	response.Body.SupportsRestartRequest = false
	response.Body.SupportsSetExpression = false
	response.Body.SupportsLoadedSourcesRequest = false
	response.Body.SupportsReadMemoryRequest = false
	response.Body.SupportsCancelRequest = false
	s.send(response)
}

func (s *Session) onLaunchRequest(request *dap.LaunchRequest) {
	// 1.check debugger
	if s.r.GetVm().GetDebugger() != nil {
		s.sendShowUserErrorResponse(request.Request, FailedToLaunch, "Failed to launch",
			fmt.Sprintf("debug session already in progress  - use remote attach mode to connect to a server with an active debug session"))
		return
	}
	// 2.parse LaunchConfig
	var args = defaultLaunchConfig // narrow copy for initializing non-zero default values
	if err := unmarshalLaunchAttachArgs(request.Arguments, &args); err != nil {
		s.sendShowUserErrorResponse(request.Request,
			FailedToLaunch, "Failed to launch", fmt.Sprintf("invalid debug configuration - %v", err))
		return
	}
	s.config.log.Debug("parsed launch config: ", prettyPrint(args))
	// 3.change working dir and env
	if args.DlvCwd != "" {
		if err := os.Chdir(args.DlvCwd); err != nil {
			s.sendShowUserErrorResponse(request.Request,
				FailedToLaunch, "Failed to launch", fmt.Sprintf("failed to chdir to %q - %v", args.DlvCwd, err))
			return
		}
	}
	for k, v := range args.Env {
		if v != nil {
			if err := os.Setenv(k, *v); err != nil {
				s.sendShowUserErrorResponse(request.Request, FailedToLaunch, "Failed to launch", fmt.Sprintf("failed to setenv(%v) - %v", k, err))
				return
			}
		} else {
			if err := os.Unsetenv(k); err != nil {
				s.sendShowUserErrorResponse(request.Request, FailedToLaunch, "Failed to launch", fmt.Sprintf("failed to unsetenv(%v) - %v", k, err))
				return
			}
		}
	}
	// 4.check config (only support debug currently)
	if args.Mode == "" {
		args.Mode = "debug"
	}
	if args.Mode != "debug" {
		s.sendShowUserErrorResponse(request.Request, FailedToLaunch, "Failed to launch",
			fmt.Sprintf("invalid debug configuration - unsupported 'mode' attribute %q", args.Mode))
		return
	}
	var err error
	// 5.build js file
	filename := args.Program
	if args.Mode == "debug" { //  || args.Mode == "test"
		src, err := os.ReadFile(filename)
		if err != nil {
			s.sendShowUserErrorResponse(request.Request, FailedToLaunch, "Failed to launch",
				fmt.Sprintf("cannot read js file,err = %v", err))
			return
		}
		prg, err := goscript.Compile(filename, string(src), false)
		if err != nil {
			s.sendShowUserErrorResponse(request.Request, FailedToLaunch, "Failed to launch",
				fmt.Sprintf("Compile err = %v", err))
			return
		}
		s.prg = prg
	}
	// 6.start if noDebug
	if args.NoDebug {
		// Skip 'initialized' event, which will prevent the client from sending
		// debug-related requests.
		s.send(&dap.LaunchResponse{Response: *newResponse(request.Request)})

		// Start the program on a different goroutine, so we can listen for disconnect request.
		go func() {
			if _, err := s.r.RunProgram(s.prg); err != nil {
				s.config.log.Debugf("program exited with error: %v", err)
			}
			//close(s.noDebugProcess.exited)
			//s.logToConsole(proc.ErrProcessExited{Pid: cmd.ProcessState.Pid(), Status: cmd.ProcessState.ExitCode()}.Error())
			s.send(&dap.TerminatedEvent{Event: *newEvent("terminated")})
		}()
		return
	}
	// 7.init debugger
	func() {
		s.mu.Lock()
		defer s.mu.Unlock() // Make sure to unlock in case of panic that will become internal error

		//s.debugger, err = debugger.New(&s.config.Debugger, s.config.ProcessArgs)
		s.r.AttachDebugger()
	}()
	if err != nil {
		s.sendShowUserErrorResponse(request.Request, FailedToLaunch, "Failed to launch", err.Error())
		return
	}
	// Notify the client that the debugger is ready to start accepting
	// configuration requests for setting breakpoints, etc. The client
	// will end the configuration sequence with 'configurationDone'.
	s.send(&dap.InitializedEvent{Event: *newEvent("initialized")})
	s.send(&dap.LaunchResponse{Response: *newResponse(request.Request)})
}

// onAttachRequest handles 'attach' request.
// This is a mandatory request to support.
// Attach debug sessions support the following modes:
//
//   - [DEFAULT] "local" -- attaches debugger to a local running process.
//     Required args: processID
//   - "remote" -- attaches client to a debugger already attached to a process.
//     Required args: none (host/port are used externally to connect)
func (s *Session) onAttachRequest(request *dap.AttachRequest) {
	// 没理解做啥用
}

func (s *Session) onDisconnectRequest(request *dap.DisconnectRequest) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.send(&dap.DisconnectResponse{Response: *newResponse(request.Request)})
	s.send(&dap.TerminatedEvent{Event: *newEvent("terminated")})
	s.conn.Close()

	s.r = nil
	s.prg = nil
}

func (s *Session) onPauseRequest(request *dap.PauseRequest) {
	s.r.Interrupt("pause")
	s.send(&dap.PauseResponse{Response: *newResponse(request.Request)})
	// No need to send any event here.
	// If we received this request while stopped, there already was an event for the stop.
	// If we received this while running, then doCommand will unblock and trigger the right
	// event, using debugger.StopReason because manual stop reason always wins even if we
	// simultaneously receive a manual stop request and hit a breakpoint.
}

func (s *Session) onContinueRequest(request *dap.ContinueRequest) {
	s.send(&dap.ContinueResponse{
		Response: *newResponse(request.Request),
		Body:     dap.ContinueResponseBody{AllThreadsContinued: true}})

	stopReason := s.r.GetVm().GetDebugger().Continue()

	if s.r.GetVm().Halted() {
		s.send(&dap.TerminatedEvent{Event: *newEvent("terminated")})
		return
	}

	stopped := &dap.StoppedEvent{Event: *newEvent("stopped")}
	stopped.Body.AllThreadsStopped = true
	stopped.Body.Reason = string(stopReason)
	s.send(stopped)
}

func (s *Session) onNextRequest(request *dap.NextRequest) {
	s.sendStepResponse(request.Arguments.ThreadId, &dap.NextResponse{Response: *newResponse(request.Request)})

	err := s.r.GetVm().GetDebugger().Next()
	if err != nil {
		s.config.log.Errorf("Error next: %v", err)
		// If we encounter an error, we will have to send a stopped event
		// since we already sent the step response.
		stopped := &dap.StoppedEvent{Event: *newEvent("stopped")}
		stopped.Body.AllThreadsStopped = true
		stopped.Body.Reason = "error"
		stopped.Body.Text = err.Error()
		s.send(stopped)
		return
	}
	// todo response stopped at where
}

func (s *Session) onStepInRequest(request *dap.StepInRequest) {
	s.sendStepResponse(request.Arguments.ThreadId, &dap.StepInResponse{Response: *newResponse(request.Request)})

	if err := s.r.GetVm().GetDebugger().StepIn(); err != nil {
		s.config.log.Errorf("Error next: %v", err)
		// If we encounter an error, we will have to send a stopped event
		// since we already sent the step response.
		stopped := &dap.StoppedEvent{Event: *newEvent("stopped")}
		stopped.Body.AllThreadsStopped = true
		stopped.Body.Reason = "error"
		stopped.Body.Text = err.Error()
		s.send(stopped)
		return
	}
}

func (s *Session) onStepOutRequest(request *dap.StepOutRequest) {
	s.sendStepResponse(request.Arguments.ThreadId, &dap.StepOutResponse{Response: *newResponse(request.Request)})

	//todo stepOut?
}

func (s *Session) sendStepResponse(threadId int, message dap.Message) {
	// All of the threads will be continued by this request, so we need to send
	// a continued event so the UI can properly reflect the current state.
	s.send(&dap.ContinuedEvent{
		Event: *newEvent("continued"),
		Body: dap.ContinuedEventBody{
			ThreadId:            threadId,
			AllThreadsContinued: true,
		},
	})
	s.send(message)
}

func (s *Session) onSetBreakpointsRequest(request *dap.SetBreakpointsRequest) {
	fileName := request.Arguments.Source.Name
	if fileName == "" {
		s.sendErrorResponse(request.Request, UnableToSetBreakpoints, "Unable to set or clear breakpoints", "empty file path")
		return
	}
	debugger := s.r.GetVm().GetDebugger()
	bps, _ := debugger.Breakpoints()
	// in dap,only setBreakpoints,no deleteBreakpoints
	// clear breakpoints
	for key := range bps {
		delete(bps, key)
	}
	path := request.Arguments.Source.Path

	breakpoints := make([]dap.Breakpoint, 0)
	// set breakpoints
	for _, breakpoint := range request.Arguments.Breakpoints {
		debugger.SetBreakpoint(fileName, breakpoint.Line)

		b := dap.Breakpoint{Line: breakpoint.Line,
			Source: &dap.Source{Name: filepath.Base(path), Path: path}}
		breakpoints = append(breakpoints, b)
	}

	response := &dap.SetBreakpointsResponse{Response: *newResponse(request.Request)}
	response.Body.Breakpoints = breakpoints
	s.send(response)
}

func (s *Session) onThreadsRequest(request *dap.ThreadsRequest) {
	var threads []dap.Thread
	// if process not start
	threads = []dap.Thread{{Id: 1, Name: "Dummy"}}

	// todo fetch thread and goroutine
	response := &dap.ThreadsResponse{
		Response: *newResponse(request.Request),
		Body:     dap.ThreadsResponseBody{Threads: threads},
	}
	s.send(response)
}

func (s *Session) onStackTraceRequest(request *dap.StackTraceRequest) {
	if s.r.GetVm().GetDebugger() == nil {
		s.sendErrorResponse(request.Request, UnableToProduceStackTrace, "Unable to produce stack trace", "debugger is nil")
		return
	}
	stackFrames := []dap.StackFrame{} // initialize to empty, since nil is not an accepted response.

	locations := s.prg.Stacktrace()
	totalFrames := len(locations)

	for _, location := range locations {
		stackFrame := dap.StackFrame{Id: int(location.PC), Line: location.Line, Name: location.FnName, InstructionPointerReference: fmt.Sprintf("%#x", location.PC)}
		stackFrames = append(stackFrames, stackFrame)
	}
	response := &dap.StackTraceResponse{
		Response: *newResponse(request.Request),
		Body:     dap.StackTraceResponseBody{StackFrames: stackFrames, TotalFrames: totalFrames},
	}
	s.send(response)
}

func (s *Session) onScopesRequest(request *dap.ScopesRequest) {
	//request.Arguments.FrameId
	// todo find stackFrame by frameId, then Retrieve function/arguments/local variables

	scopes := []dap.Scope{}

	response := &dap.ScopesResponse{
		Response: *newResponse(request.Request),
		Body:     dap.ScopesResponseBody{Scopes: scopes},
	}
	s.send(response)
}

func (s *Session) onVariablesRequest(request *dap.VariablesRequest) {
	debugger := s.r.GetVm().GetDebugger()
	if debugger == nil {
		s.sendErrorResponseWithOpts(request.Request, UnableToLookupVariable, "Unable to lookup variable", "debugger is nil", true)
		return
	}
	variables, err := debugger.GetGlobalVariables()
	if err != nil {
		s.sendErrorResponse(request.Request, UnableToLookupVariable, "Unable to lookup variable", err.Error())
		return
	}
	children := []dap.Variable{} // must return empty array, not null, if no children
	for name, value := range variables {
		v := dap.Variable{Name: name,
			Value: value.String(),
			Type:  value.ExportType().String()}
		children = append(children, v)
	}

	response := &dap.VariablesResponse{
		Response: *newResponse(request.Request),
		Body:     dap.VariablesResponseBody{Variables: children},
	}
	s.send(response)
}

func (s *Session) onEvaluateRequest(request *dap.EvaluateRequest) {
	debugger := s.r.GetVm().GetDebugger()
	if debugger == nil {
		s.sendErrorResponseWithOpts(request.Request, UnableToEvaluateExpression, "Unable to evaluate expression", "debugger is nil", true)
		return
	}
}

func newResponse(request dap.Request) *dap.Response {
	return &dap.Response{
		ProtocolMessage: dap.ProtocolMessage{
			Seq:  0,
			Type: "response",
		},
		Command:    request.Command,
		RequestSeq: request.Seq,
		Success:    true,
	}
}

func newEvent(event string) *dap.Event {
	return &dap.Event{
		ProtocolMessage: dap.ProtocolMessage{
			Seq:  0,
			Type: "event",
		},
		Event: event,
	}
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
