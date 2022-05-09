package apiclient

import (
	"fmt"
	"net"
	"net/http"
	"net/url"

	"github.com/google/uuid"
	"github.com/pkg/browser"
)

// AuthClient represents a client for Infracost's authentication process.
type AuthClient struct {
	Host string
}

// Login opens a browser with authentication URL and starts a HTTP server to
// wait for a callback request.
func (a AuthClient) Login() (string, error) {
	state := uuid.NewString()

	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return "", err
	}
	port := listener.Addr().(*net.TCPAddr).Port

	q := url.Values{}
	q.Set("port", fmt.Sprint(port))
	q.Set("state", state)

	startURL := fmt.Sprintf("%s/login?%s", a.Host, q.Encode())

	fmt.Println("Opening the authentication URL in your browser.")
	fmt.Println("\nIf not opened automatically, copy and paste it to your browser:")
	fmt.Printf("\n    %s\n\n", startURL)
	fmt.Println("Waiting...")

	_ = browser.OpenURL(startURL)

	apiKey, err := a.startCallbackServer(listener, state)
	if err != nil {
		return "", err
	}

	return apiKey, nil
}

func (a AuthClient) startCallbackServer(listener net.Listener, generatedState string) (string, error) {
	apiKey := ""

	err := http.Serve(listener, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		defer listener.Close()

		query := r.URL.Query()
		state := query.Get("state")
		apiKey = query.Get("apiKey")

		if apiKey == "" || state != generatedState {
			w.WriteHeader(400)
			return
		}
	}))

	if err != nil {
		return "", err
	}

	return apiKey, nil
}
