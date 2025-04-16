package gateway_test

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/cloud-gov/billing/internal/gateway"
)

// CheckInternalErr should be called on any error that indicates a problem with the test itself if it is not nil.
func CheckInternalErr(t *testing.T, err error) {
	if err != nil {
		t.Fatal("internal error occurred; this indicates a problem with the test:", err)
	}
}

func TestHandleRequest(t *testing.T) {
	dest := httptest.NewServer(http.HandlerFunc(HandleDestination))
	defer dest.Close()

	u, err := url.Parse(dest.URL)
	CheckInternalErr(t, err)
	gateway := httptest.NewServer(http.HandlerFunc(gateway.HandleRequest(u.Host, slog.Default())))
	defer gateway.Close()

	req := httptest.NewRequest("GET", gateway.URL, strings.NewReader(""))
	resp, err := http.DefaultTransport.RoundTrip(req)
	if err != nil {
		t.Fatal("making the request to the proxy server: ", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal("reading response body from proxy server", err)
	}
	if string(body) != "Received!" {
		t.Fatal("unexpected response body from proxy server")
	}
}

func TestHandleRequest_DestHost(t *testing.T) {
	logger := slog.Default()
	rec := httptest.NewRecorder()

	cases := []struct {
		Name               string
		DestHost           string
		ExpectedStatusCode int
	}{
		{
			Name:               "host only",
			DestHost:           "example.gov",
			ExpectedStatusCode: http.StatusOK,
		},
		{
			Name:               "host:port",
			DestHost:           "example.gov:8765",
			ExpectedStatusCode: http.StatusOK,
		},
		{
			Name:               "empty",
			DestHost:           "",
			ExpectedStatusCode: http.StatusBadGateway,
		},
	}
	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			h := gateway.HandleRequest(tc.DestHost, logger)
			h.ServeHTTP(rec, httptest.NewRequest("GET", "http://example.invalid", strings.NewReader("")))
			t.Log(rec.Result())
			if rec.Result().StatusCode != http.StatusOK {
				t.Fatalf("expected response code %v, got %v", http.StatusOK, rec.Result().StatusCode)
			}
		})
	}
}

func HandleDestination(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Received!")
}
