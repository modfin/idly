package idly

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

func NewClient(serviceName, uri string) *Client {
	return &Client{
		uri:     uri,
		service: serviceName,
	}
}

type Client struct {
	uri     string
	service string
}

type Request struct {
	client *Client
	login  Login
}

// List successful logins of a user of a service, records
func (c *Client) List(uid string) ([]Login, error) {
	res, err := http.Get(fmt.Sprintf("%s/login/%s/%s", c.uri, url.PathEscape(c.service), url.PathEscape(uid)))
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	var logins []Login
	err = json.NewDecoder(res.Body).Decode(&logins)
	return logins, err
}

func (c *Client) Request(email string, ipAddress string) *Request {
	return &Request{
		client: c,
		login: Login{
			Service:     c.service,
			UID:         email,
			At:          time.Now(),
			Email:       email,
			IPAddress:   ipAddress,
			HttpHeaders: map[string][]string{},
		}}
}

func Headers(h http.Header) func(request *Request) {
	return func(request *Request) {
		for k, v := range h {
			request.login.HttpHeaders[k] = v
		}
	}
}
func UserAgent(ua string) func(request *Request) {
	return func(request *Request) {
		request.login.HttpHeaders.Set("User-Agent", ua)
	}
}
func UserId(uid string) func(request *Request) {
	return func(request *Request) {
		request.login.UID = uid
	}
}

func (r *Request) With(ops ...func(request *Request)) *Request {
	for _, o := range ops {
		o(r)
	}
	return r
}

func (r *Request) WithHeaders(h http.Header) *Request {
	r.With(Headers(h))
	return r
}
func (r *Request) WithUserId(uid string) *Request {
	r.With(UserId(uid))
	return r
}
func (r *Request) WithUserAgent(ua string) *Request {
	r.With(UserAgent(ua))
	return r
}

// Success reports a successgful login too idly. This is done async.
//
//	This is used in ordern to figure out if a Warning mail shall be sent or not
func (r *Request) Success() {
	data, err := json.Marshal(r.login)
	if err != nil {
		fmt.Println("[Idly Client] could not marshal login struct")
	}
	go func() {
		_, err = http.Post(fmt.Sprintf("%s/login", r.client.uri), "application/json", bytes.NewBuffer(data))
		if err != nil {
			fmt.Println("[Idly Client] could not make login request; err:", err)
		}
	}()
}

// Fail reports a failed login attempt too idly. This is done async.
//
//	This is used by idly to collect metrics which can be monitored through prometheus
func (r *Request) Fail() {
	data, err := json.Marshal(r.login)
	if err != nil {
		fmt.Println("[Idly Client] could not marshal login struct")
	}
	go func() {
		_, err = http.Post(fmt.Sprintf("%s/login-fail", r.client.uri), "application/json", bytes.NewBuffer(data))
		if err != nil {
			fmt.Println("[Idly Client] could not make login-fail request; err:", err)
		}
	}()
}
