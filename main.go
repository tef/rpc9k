package main

import (
	"fmt"

	"github.com/tef/rpc9k/wire"
	"github.com/tef/rpc9k/client"
)

func main() {
	c2 := &client.Client{
		Url:"/",
		Options: nil,
		Message: wire.Root,
		Err: nil,
	}

	r := c2.Fetch("Example")

	fmt.Println("fetch", r)
	fmt.Println("begin")

	c := client.Dial("url", &client.Auth{Name: "n", Token: "t"})
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
