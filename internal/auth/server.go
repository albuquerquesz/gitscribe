package auth

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"
)


type CallbackServer struct {
	port       int
	server     *http.Server
	resultChan chan *CallbackResult
	state      string
	stateMu    sync.RWMutex
}


type CallbackResult struct {
	Code  string
	State string
	Error string
}


func NewCallbackServer(preferredPort int) (*CallbackServer, int, error) {
	
	ports := append([]int{preferredPort}, AlternativePorts...)

	var listener net.Listener
	var selectedPort int

	for _, port := range ports {
		addr := fmt.Sprintf("localhost:%d", port)
		l, err := net.Listen("tcp", addr)
		if err == nil {
			listener = l
			selectedPort = port
			break
		}
	}

	if listener == nil {
		return nil, 0, ErrPortInUse
	}

	resultChan := make(chan *CallbackResult, 1)

	server := &http.Server{
		Addr:         listener.Addr().String(),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	cs := &CallbackServer{
		port:       selectedPort,
		server:     server,
		resultChan: resultChan,
	}

	
	mux := http.NewServeMux()
	mux.HandleFunc("/callback", cs.handleCallback)
	mux.HandleFunc("/health", cs.handleHealth)
	server.Handler = mux

	
	go func() {
		if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
			
			fmt.Printf("Callback server error: %v\n", err)
		}
	}()

	return cs, selectedPort, nil
}


func (cs *CallbackServer) SetState(state string) {
	cs.stateMu.Lock()
	defer cs.stateMu.Unlock()
	cs.state = state
}


func (cs *CallbackServer) GetState() string {
	cs.stateMu.RLock()
	defer cs.stateMu.RUnlock()
	return cs.state
}


func (cs *CallbackServer) handleCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")
	errorParam := r.URL.Query().Get("error")

	if errorParam != "" {
		cs.resultChan <- &CallbackResult{
			Error: errorParam,
		}
		http.Error(w, fmt.Sprintf("OAuth error: %s", errorParam), http.StatusBadRequest)
		return
	}

	
	expectedState := cs.GetState()
	if state == "" || state != expectedState {
		cs.resultChan <- &CallbackResult{
			Error: "invalid_state",
		}
		http.Error(w, "Invalid state parameter", http.StatusBadRequest)
		return
	}

	cs.resultChan <- &CallbackResult{
		Code:  code,
		State: state,
	}

	
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(successHTML))
}


func (cs *CallbackServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"healthy"}`))
}


func (cs *CallbackServer) WaitForCallback(ctx context.Context) (*CallbackResult, error) {
	select {
	case result := <-cs.resultChan:
		return result, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}


func (cs *CallbackServer) Stop(ctx context.Context) error {
	return cs.server.Shutdown(ctx)
}


const successHTML = `<!DOCTYPE html>
<html>
<head>
    <title>Authentication Successful</title>
    <style>
        body { 
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            display: flex; 
            justify-content: center; 
            align-items: center; 
            height: 100vh; 
            margin: 0; 
            background: #f5f5f5;
        }
        .container { 
            text-align: center; 
            background: white;
            padding: 40px;
            border-radius: 8px;
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
        }
        .checkmark { 
            color: #10a37f; 
            font-size: 64px; 
            margin-bottom: 20px;
        }
        h1 { color: #202123; margin-bottom: 10px; }
        p { color: #6e6e80; }
    </style>
</head>
<body>
    <div class="container">
        <div class="checkmark">✓</div>
        <h1>Authentication Successful</h1>
        <p>You can close this window and return to the CLI.</p>
    </div>
</body>
</html>`
