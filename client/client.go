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
	Options  any
	Url      string
	Response wire.Variant
	Err      error
	Cache    map[string]*Client
}

func (c *Client) Variant() wire.Variant {
	if c.Response.IsEmpty() {
		return wire.Variant{} // Maybe put an empty in there but whatever
	}
	return c.Response
}

func New(rawUrl string, variant wire.Variant, options any) *Client {
	return &Client{
		Options:  options,
		Url:      rawUrl,
		Response: variant,
		Err:      nil,
		Cache:    make(map[string]*Client),
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

func (c *Client) withErr(err error) *Client {
	return &Client{
		Response: wire.NewErr("error", err),
		Options:  c.Options,
		Err:      err,
	}
}

func (c *Client) withNewError(args ...any) *Client {
	text := fmt.Sprintln(args...)
	err := errors.New(text)
	return &Client{
		Response: wire.NewErr("error", err),
		Options:  c.Options,
		Err:      err,
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
	} else if c.Response.IsEmpty() || c.Url == "" {
		return c.withNewError("No url opened")
	}
	return c.Fetch(name).Call(args)
}

func (c *Client) Call(args any) *Client {
	if c.Err != nil {
		return c
	} else if c.Response.IsEmpty() || c.Url == "" {
		return c.withNewError("No url opened")
	}
	var env wire.Variant
	if e, ok := args.(wire.Variant); ok {
		env = e
	} else if m, ok := args.(wire.WireMessage); ok {
		env = m.Variant()
	} else { // Raw Go Value
		value := &wire.Value{Value: args}
		env = value.Variant()
	}

	request := c.Response.Call(env, c.Url)
	return c.Request(request)

}

func (c *Client) Fetch(path string) *Client {
	if c.Err != nil {
		return c
	} else if c.Response.IsEmpty() || c.Url == "" {
		return c.withNewError("No url opened")
	}
	if c.Cache != nil {
		if client, ok := c.Cache[path]; ok {
			// fmt.Println("Cached Fetch:", path)
			return client
		}
	}

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

		request := client.Response.Fetch(name, client.Url)

		// fmt.Println("Request:", prefix, request)

		if request == nil {
			client = c.withNewError("can't fetch:", name)
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

	if r == nil {
		c.withNewError("empty (nil) request")
	}

	url := c.urlFor(r)

	if r.Cached != nil {
		client := &Client{
			Response: r.Cached.Variant(),
			Url:      url,
			Options:  c.Options,
			Err:      nil,
		}
		return client
	}

	var variant *wire.Variant
	var err error

	for {
		url, variant, err = c.httpRequest(url, r)
		if err != nil {
			return c.withErr(err)
		}

		// future handling
		// don't update url, but fetch next in series
		// waiting if necessary

		break
	}

	fmt.Println("Variant reply:", variant.Kind)

	client := &Client{
		Response: *variant,
		Url:      url,
		Options:  c.Options,
	}
	return client

}

func (c *Client) httpRequest(url string, r *wire.Request) (string, *wire.Variant, error) {

	payload, err := r.Body()
	if err != nil {
		return url, nil, err
	}

	variant, err := wire.FakeServer(r.Action, url, payload)
	// redirect handling/ faking
	// we should get blob back from FakeServer, and promote
	// it to an variant with a JSON if the contents are JSON
	// and a Message if the contents are the message conten/type

	if err != nil {
		return url, nil, err
	}

	if variant.Kind == "Redirect" {
		redirect, ok := variant.Msg.(*wire.Redirect)
		if ok {
			url = redirect.Url(url)
			fmt.Println("fetching redirected", url)
			_, variant, err = c.httpRequest(url, r)

			if err != nil {
				return url, nil, err
			}
		}

	}

	return url, variant, nil

}
func (c *Client) Blob() (*wire.Blob, error) {
	if c.Err != nil {
		return nil, c.Err
	}
	if c.Response.IsEmpty() {
		return nil, nil
	}

	return c.Response.Blob()
}

func (c *Client) Scan(out any) *Client {
	if c.Err != nil {
		return c
	}

	if c.Response.IsEmpty() && out == nil {
		return c
	} else if c.Response.IsEmpty() || out == nil {
		return c.withNewError("bad. nil")
	}

	err := c.Response.Scan(out)
	if err != nil {
		return c.withErr(err)
	}
	return c
}
