package main

import (
	"errors"
	"fmt"
	"time"
)

type Message interface {
	Kind() string
	Routes() []string
	Fetch(name string) (*Request, Message)
	Call(args any) (*Request)
}

type Auth struct {
	Name  string
	Token string
}

type Request struct {
	Action string // adverb: get, call, list
	Path   string
	State  any
	Args   any
}

func (r *Request) UrlFrom(base string) string {
	return base+","+r.Path
}
type Client struct {
	Options any
	Url string
	Message Message
	Err     error
}

func Dial(url string, options any) *Client {
	fmt.Println("open", url)

	var message Message
	message = &Namespace{}
	client := &Client{
		Url: url,
		Message: message,
		Options: options,
		Err:     nil,
	}
	return client
}

func (c *Client) Request(r *Request) *Client {
	// build a httprequest(method, url) from Request

	// if reply is a future:
	// while True
	// reply := http.Blah
	//     ...
	// decode args
	// if json, wrap in Result
	// if 9k type, bring it up

	client := &Client{
		Message: c.Message,
		Url: c.Url,
		Options: c.Options,
		Err:     nil,
	}
	return client

}

func (c *Client) Invoke(name string, args any) *Client {
	if c.Err != nil {
		return c
	}
	api := c.Fetch(name)
	return api.Call(args)
}

func (c *Client) Call(args any) *Client {
	if c.Err != nil {
		return c
	}

	request := c.Message.Call(args)
	return c.Request(request)

}

func (c *Client) Fetch(name string) *Client {
	if c.Err != nil {
		return c
	}
	request, message := c.Message.Fetch(name)

	if message != nil {
		return &Client{
			Options: c.Options,
			Url: c.Url,
			Message: message,
			Err:     nil,
		}
	} else if request == nil {
		return &Client{
			Options: c.Options,
			Url: "",
			Message: nil,
			Err: errors.New("can't fetch "+name),
		}
	}

	return c.Request(request)
}

func (c *Client) Scan(out any) *Client {
	return c
}

type Metadata struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	Version   int
}

type CommonMessage struct {
	kind       string
	ApiVersion string
	Metadata   Metadata
}

func (b *CommonMessage) Kind() string {
	return b.kind
}

func (b *CommonMessage) Routes() []string {
	return []string{}
}

func (b *CommonMessage) Fetch(name string) (*Request, Message) {
	return nil, nil
}

func (b *CommonMessage) Call(args any) *Request {
	return nil
}

type Namespace struct {
	CommonMessage
	routes []string
	urls   map[string]string
	embeds map[string]Message
}

func (n *Namespace) Kind() string {
	return "Namespace"
}

func (n *Namespace) Routes() []string {
	return n.routes
}

func (n *Namespace) Fetch(name string) (*Request, Message) {
	request := Request{
		Action:"get",
		Path: "",
	}
	url, ok := n.urls[name]
	if ok {
		request.Path = url
	} else {
		request.Path = name
	}

	message, ok := n.embeds[name]

	if ok {
		return &request, message
	}

	return &request, nil
}

type Service struct {
	CommonMessage
}

func (*Service) Kind() string {
	return "Service"
}

type Procedure struct {
	CommonMessage
}

func (*Procedure) Kind() string {
	return "Procedure"
}

var root = &Namespace{
	CommonMessage: CommonMessage{
		kind: "Namespace",
		ApiVersion: "0",
	},
	routes: []string{"Example",},
	urls: map[string]string{},
	embeds: map[string]Message{},
}


func main() {
	c2 := &Client{
		Url:"/",
		Options: nil,
		Message: root,
		Err: nil,
	}

	r := c2.Fetch("Example")

	fmt.Println("fetch", r)
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
