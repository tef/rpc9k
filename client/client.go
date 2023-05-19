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
}

func Dial(url string, options any) *Client {
	fmt.Println("open", url)

	var message wire.Message
	message = &wire.Namespace{}
	client := &Client{
		Url: url,
		Message: message,
		Options: options,
		Err:     nil,
	}
	return client
}

func (c *Client) Request(r *wire.Request) *Client {
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

