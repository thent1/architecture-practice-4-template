package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"gopkg.in/check.v1"
)

type BalancerSuite struct{}

var _ = check.Suite(&BalancerSuite{})

func Test(t *testing.T) {
	check.TestingT(t)
}

func (s *BalancerSuite) TestHealth(c *check.C) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	serversPool = []string{server.Listener.Addr().String()}
	healthyServers = make([]bool, len(serversPool))

	isHealthy := health(serversPool[0])
	c.Assert(isHealthy, check.Equals, true)
}

func (s *BalancerSuite) TestChooseServer(c *check.C) {
	serversPool = []string{"server1:8080", "server2:8080", "server3:8080"}
	healthyServers = []bool{false, true, true}

	server := chooseServer("/testpath")
	c.Assert(server, check.Equals, "server2:8080")

	healthyServers = []bool{false, false, false}
	server = chooseServer("/testpath")
	c.Assert(server, check.Equals, "")
}

func (s *BalancerSuite) TestForward(c *check.C) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))
	defer backend.Close()

	serversPool = []string{backend.Listener.Addr().String()}
	healthyServers = []bool{true}

	req := httptest.NewRequest("GET", "http://localhost/test", nil)
	rw := httptest.NewRecorder()

	err := forward(serversPool[0], rw, req)
	c.Assert(err, check.IsNil)
	c.Assert(rw.Code, check.Equals, http.StatusOK)
	c.Assert(rw.Body.String(), check.Equals, "OK")
}

func (s *BalancerSuite) TestMain(c *check.C) {

	go main()

	time.Sleep(1 * time.Second)

	resp, err := http.Get("http://localhost:8090/health")
	c.Assert(err, check.IsNil)
	c.Assert(resp.StatusCode, check.Equals, http.StatusServiceUnavailable)

}
