package main

import (
	"fmt"
)

type Message interface {
	Kind() string
	Routes() []string
	Link(name string) string
	Cached(name string) Message
}

type Auth struct {
	Name  string
	Token string
}

type Request struct {
	Method string
	Path   string
	Args   any
}

type Client struct {
	BaseUrl string
	Message Message
	Options any
	Err     error
}

func Dial(url string, options any) *Client {
	fmt.Println("open", url)

	var message Message
	message = &Namespace{}
	client := &Client{
		BaseUrl: url,
		Message: message,
		Options: options,
		Err:     nil,
	}
	return client
}
func (c *Client) cacheResult(url string, m Message) {
	return
}

func (c *Client) cachedResult(url string) Message {
	return nil
}

func (c *Client) addressOf(name string) string {
	// if name isn't resolved, look things up
	// as you go
	// saving the result

	// split up name by "."s
	// for each prefix:
	// 	m.Routes()
	//	m.Link() & caching prefix address
	//	m.Embed() & caching Message at address
	// 	request(addr) & caching Message at address
	// for tail
	//	m.Routes() & m.Link() & caching full url
	//	m.Embed() & caching embed
	//	return addr
	return "/foo"
}

func (c *Client) request(r *Request) *Client {
	// encode args as json
	//
	// if reply is a future:
	// while True
	// reply := http.Blah
	//     ...
	// decode args
	// if json, wrap in Result
	// if 9k type, bring it up

	client := &Client{
		Message: c.Message,
		BaseUrl: c.BaseUrl,
		Options: c.Options,
		Err:     nil,
	}
	return client

}

func (c *Client) Invoke(name string, args any) *Client {
	if c.Err != nil {
		return c
	}
	url := c.addressOf(name)
	request := &Request{
		Method: "POST",
		Path:   url,
		Args:   args,
	}

	return c.request(request)
}

func (c *Client) Fetch(name string) *Client {
	if c.Err != nil {
		return c
	}
	url := c.addressOf(name)

	message := c.cachedResult(url)

	if message != nil {
		return &Client{
			BaseUrl: url,
			Message: message,
			Err:     nil,
			Options: c.Options,
		}
	}

	request := &Request{
		Method: "GET",
		Path:   url,
		Args:   nil,
	}
	result := c.request(request)
	if result.Err != nil {
		c.cacheResult(url, result.Message)
	}
	return result
}

func (c *Client) Scan(out any) *Client {
	return c
}

type Metadata struct {
}

type BaseMessage struct {
	kind       string
	ApiVersion string
	Metadata   Metadata
}

func (b *BaseMessage) Kind() string {
	return b.kind
}

func (b *BaseMessage) Routes() []string {
	return []string{}
}

func (b *BaseMessage) Link(name string) string {
	return ""
}

func (b *BaseMessage) Cached(name string) Message {
	return nil
}

type Namespace struct {
	BaseMessage
}

func (*Namespace) Kind() string {
	return "Namespace"
}

type Service struct {
	BaseMessage
}

func (*Service) Kind() string {
	return "Service"
}

type Procedure struct {
	BaseMessage
}

func (*Procedure) Kind() string {
	return "Procedure"
}

func main() {
	fmt.Println("begin")

	c := Dial("url", &Auth{Name: "n", Token: "t"})
	if c.Err != nil {
		fmt.Println("err:", c.Err)
		return
	}

	var Output struct {
		A int
		B int
	}

	reply := c.Invoke("Example:rpc", []int{1, 2, 3}).Scan(&Output)

	if reply.Err != nil {
		fmt.Println("err:", reply.Err)
		return
	}

	fmt.Println("Output", Output)

	example := c.Fetch("Example")
	if example.Err != nil {
		fmt.Println("err:", example.Err)
		return
	}

	reply = example.Invoke("rpc", []int{1, 2, 3}).Scan(&Output)
	if reply.Err != nil {
		fmt.Println("err:", reply.Err)
		return
	}

	fmt.Println("Output", Output)

	fmt.Println("end")

}
