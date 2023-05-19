package client

import (
	"fmt"
	"errors"
	neturl "net/url"

	"github.com/tef/rpc9k/wire"
)


type Auth struct {
	Name  string
	Token string
}

type Client struct {
	Options any
	Url string
	Message wire.Message
	Err     error
	Cache map[string]*Client
}

func New(rawUrl string, message wire.Message, options any) *Client{
	return &Client{
		Options: options,
		Url: rawUrl,
		Message: message,
		Err: nil,
		Cache: make(map[string]*Client),
	}
}

func Dial(rawUrl string, options any) *Client {
	request := &wire.Request{
		Action: "get",
		Path: rawUrl,
	}

	client := &Client{
		Options: options,
		Url: "",
	}
	return client.Request(request)
}

func (c *Client) SetErr(err error) *Client {
	return &Client{
		Options: c.Options,
		Err: err,
	}
}

func (c *Client) Invoke(name string, args any) *Client {
	if c.Err != nil {
		return c
	}
	return c.Fetch(name).Call(args)
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
	if c.Cache != nil {
		if client, ok := c.Cache[name]; ok {
			fmt.Println("Cached Fetch:", name)
			return client
		}
	}
	// var client = c
	// for parts in name.split ":"
	//	req = c.Message.Fetch(name)
	//	client = c.Request(req)

	fmt.Println("Non Cached Fetch:", name)
			
	client := c
	request := client.Message.Fetch(name)

	fmt.Println("Request:", request)

	if request == nil {
		client =  c.SetErr(errors.New("can't fetch "+name))
	} else {
		client = c.Request(request)

		if client.Err == nil && c.Cache != nil {
			c.Cache[name] = client
		}
	}
	return client
}

func (c *Client) Request(r *wire.Request) *Client {
	if c.Err != nil {
		return c
	}

	if r != nil && r.Cached != nil {
		client := &Client{
			Message: r.Cached,
			Url: c.joinUrl(r),
			Options: c.Options,
			Err: nil,
		}
		return client
	}
	// build a httprequest(method, url) from Request

	// if reply is a future:
	// while True
	// reply := http.Blah
	//     ...
	// decode args
	// if json, wrap in Result
	// if 9k type, bring it up

	client := &Client{
		Message: wire.Root,
		Url: c.joinUrl(r),
		Options: c.Options,
		Err:     nil,
	}
	return client

}

func (c *Client) joinUrl(r *wire.Request) string{
	if r == nil {
		return c.Url
	}
	if c.Url == "" {
		return r.Path
	}

	url, _:= neturl.JoinPath(c.Url, r.Path)
	return url
}

func (c *Client) Scan(out any) *Client {
	if c.Err != nil {
		return c
	}

	return c
}

