package client

import (
	"errors"
	"fmt"
	//neturl "net/url"
	"strings"

	"github.com/tef/rpc9k/wire"
)

type Auth struct {
	Name  string
	Token string
}

type Client struct {
	Options any
	Url     string
	Message wire.Message
	Err     error
	Cache   map[string]*Client
}

func New(rawUrl string, message wire.Message, options any) *Client {
	return &Client{
		Options: options,
		Url:     rawUrl,
		Message: message,
		Err:     nil,
		Cache:   make(map[string]*Client),
	}
}

func Dial(rawUrl string, options any) *Client {
	request := &wire.Request{
		Action: "get",
		Base:   rawUrl,
	}

	client := &Client{
		Options: options,
	}
	client = client.Request(request)
	client.Cache = make(map[string]*Client)
	return client
}

func (c *Client) setErr(err error) *Client {
	return &Client{
		Message: &wire.Error{Id: "error", Text: err.Error()},
		Options: c.Options,
		Err:     err,
	}
}

func (c *Client) setErrorText(args ...any) *Client {
	text := fmt.Sprintln(args...)
	err := errors.New(text)
	return &Client{
		Message: &wire.Error{Id: "error", Text: text},
		Options: c.Options,
		Err:     err,
	}
}

func (c *Client) urlFor(r *wire.Request) string {
	if r == nil {
		return c.Url
	}
	return r.Url(c.Url)
}

func (c *Client) Invoke(name string, args any) *Client {
	if c.Err != nil {
		return c
	} else if c.Message == nil || c.Url == "" {
		return c.setErrorText("No url opened")
	}
	return c.Fetch(name).Call(args)
}

func (c *Client) Call(args any) *Client {
	if c.Err != nil {
		return c
	} else if c.Message == nil || c.Url == "" {
		return c.setErrorText("No url opened")
	}

	request := c.Message.Call(args, c.Url)
	return c.Request(request)

}

func (c *Client) Fetch(path string) *Client {
	if c.Err != nil {
		return c
	} else if c.Message == nil || c.Url == "" {
		return c.setErrorText("No url opened")
	}
	if c.Cache != nil {
		if client, ok := c.Cache[path]; ok {
			// fmt.Println("Cached Fetch:", path)
			return client
		}
	}
	// var client = c
	// for parts in name.split ":"
	//	req = c.Message.Fetch(name)
	//	client = c.Request(req)

	segments := strings.Split(path, ":")
	prefix := ""
	client := c
	for _, name := range segments {
		if prefix == "" {
			prefix = name
		} else {
			prefix += ":" + name
		}

		// fmt.Println("Non Cached Fetch:", path, "getting", prefix)

		request := client.Message.Fetch(name, client.Url)

		// fmt.Println("Request:", prefix, request)

		if request == nil {
			client = c.setErrorText("can't fetch:", name)
		} else {
			client = c.Request(request)

			if client.Err == nil && c.Cache != nil {
				c.Cache[prefix] = client
				// fmt.Println("Add Cached", c.Cache)
			}
		}
	}

	return client
}

func (c *Client) Request(r *wire.Request) *Client {
	if c.Err != nil {
		return c
	}

	url := c.urlFor(r)

	if r != nil && r.Cached != nil {
		client := &Client{
			Message: r.Cached,
			Url:     url,
			Options: c.Options,
			Err:     nil,
		}
		return client
	}

	// if reply is a future:
	// decode args
	// if json, wrap in Result
	// if 9k type, bring it up

	output, err := wire.FakeServer(url, r)

	client := &Client{
		Message: output,
		Url:     url,
		Options: c.Options,
		Err:     err,
	}
	return client

}

func (c *Client) Scan(out any) *Client {
	if c.Err != nil {
		return c
	}

	err := c.Message.Scan(out)
	if err != nil {
		return c.setErr(err)
	}
	return c
}
