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
	Envelope wire.Envelope
	Err      error
	Cache    map[string]*Client
}

func New(rawUrl string, envelope wire.Envelope, options any) *Client {
	return &Client{
		Options:  options,
		Url:      rawUrl,
		Envelope: envelope,
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
		Envelope: wire.NewErr("error", err),
		Options:  c.Options,
		Err:      err,
	}
}

func (c *Client) withNewError(args ...any) *Client {
	text := fmt.Sprintln(args...)
	err := errors.New(text)
	return &Client{
		Envelope: wire.NewErr("error", err),
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
	} else if c.Envelope.IsEmpty() || c.Url == "" {
		return c.withNewError("No url opened")
	}
	return c.Fetch(name).Call(args)
}

func (c *Client) Call(args any) *Client {
	if c.Err != nil {
		return c
	} else if c.Envelope.IsEmpty() || c.Url == "" {
		return c.withNewError("No url opened")
	}
	var env wire.Envelope
	if e, ok := args.(wire.Envelope); ok {
		env = e
	} else if m, ok := args.(wire.WireMessage); ok {
		env = m.Wrap()
	} else { // Raw Go Value
		value := &wire.Value{Value: args}
		env = value.Wrap()
	}

	request := c.Envelope.Call(env, c.Url)
	return c.Request(request)

}

func (c *Client) Fetch(path string) *Client {
	if c.Err != nil {
		return c
	} else if c.Envelope.IsEmpty() || c.Url == "" {
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

		request := client.Envelope.Fetch(name, client.Url)

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
			Envelope: r.Cached.Wrap(),
			Url:      url,
			Options:  c.Options,
			Err:      nil,
		}
		return client
	}

	var envelope *wire.Envelope
	var err error

	for {
		url, envelope, err = c.httpRequest(url, r)
		if err != nil {
			return c.withErr(err)
		}

		// future handling
		// don't update url, but fetch next in series
		// waiting if necessary

		break
	}

	fmt.Println("Envelope reply:", envelope.Kind)

	client := &Client{
		Envelope: *envelope,
		Url:      url,
		Options:  c.Options,
	}
	return client

}

func (c *Client) httpRequest(url string, r *wire.Request) (string, *wire.Envelope, error) {

	payload, err := r.Body()
	if err != nil {
		return url, nil, err
	}

	envelope, err := wire.FakeServer(r.Action, url, payload)
	// redirect handling/ faking
	// we should get blob back from FakeServer, and promote
	// it to an envelope with a JSON if the contents are JSON
	// and a Message if the contents are the message conten/type

	if err != nil {
		return url, nil, err
	}

	if envelope.Kind == "Redirect" {
		redirect, ok := envelope.Msg.(*wire.Redirect)
		if ok {
			url = redirect.Url(url)
			fmt.Println("fetching redirected", url)
			_, envelope, err = c.httpRequest(url, r)

			if err != nil {
				return url, nil, err
			}
		}

	}

	return url, envelope, nil

}
func (c *Client) Blob() (*wire.Blob, error) {
	if c.Err != nil {
		return nil, c.Err
	}
	if c.Envelope.IsEmpty() {
		return nil, nil
	}

	return c.Envelope.Blob()
}

func (c *Client) Scan(out any) *Client {
	if c.Err != nil {
		return c
	}

	if c.Envelope.IsEmpty() && out == nil {
		return c
	} else if c.Envelope.IsEmpty() || out == nil {
		return c.withNewError("bad. nil")
	}

	err := c.Envelope.Scan(out)
	if err != nil {
		return c.withErr(err)
	}
	return c
}
