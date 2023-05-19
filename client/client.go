package client

import (
	"fmt"
	"errors"

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

func New(url string, message wire.Message, options any) *Client{
	return &Client{
		Options: options,
		Url: url,
		Message: message,
		Err: nil,
		Cache: make(map[string]*Client),
	}
}

func Dial(url string, options any) *Client {

	request := &wire.Request{
		Action: "get",
		Path: url,
	}

	client := &Client{
		Url: url,
		Message: nil,
		Options: options,
		Err:     nil,
	}
	return client.Request(request)
}

func (c *Client) Error(err error) *Client {
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
			return client
		}
	}
	// var client = c
	// for parts in name.split ":"
	//	req = c.Message.Fetch(name)
	//	client = c.Request(req)

	fmt.Println("Fetch:", name)
			
	client := c
	request := client.Message.Fetch(name)

	fmt.Println("Request:", request)

	if request == nil {
		client =  &Client{
			Options: c.Options,
			Url: "",
			Message: nil,
			Err: errors.New("can't fetch "+name),
		}
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
			Url: r.UrlFrom(c.Url),
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
		Url: c.Url,
		Options: c.Options,
		Err:     nil,
	}
	return client

}


func (c *Client) Scan(out any) *Client {
	if c.Err != nil {
		return c
	}

	return c
}

