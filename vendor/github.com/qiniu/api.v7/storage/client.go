package storage

import (
	"github.com/qiniu/api.v7/auth/qbox"
	"github.com/qiniu/x/rpc.v7"
	"net/http"
)

type Transport struct {
	mac       qbox.Mac
	Transport http.RoundTripper
}

func (t *Transport) NestedObject() interface{} {
	return t.Transport
}

func (t *Transport) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	token, err := t.mac.SignRequest(req)
	if err != nil {
		return
	}
	req.Header.Set("Authorization", "QBox "+token)
	return t.Transport.RoundTrip(req)
}

func NewTransport(mac *qbox.Mac, transport http.RoundTripper) *Transport {
	if transport == nil {
		transport = http.DefaultTransport
	}
	t := &Transport{mac: *mac, Transport: transport}
	return t
}

func NewClient(mac *qbox.Mac, transport http.RoundTripper) *rpc.Client {
	t := NewTransport(mac, transport)
	return &rpc.Client{&http.Client{Transport: t}}
}
