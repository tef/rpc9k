package main

import (
	"encoding/json"
	"fmt"

	"github.com/tef/rpc9k/client"
	"github.com/tef/rpc9k/wire"
)

func main() {

	c2 := client.New("/", wire.Root, nil)
	r := c2.Fetch("Example")

	if r.Err != nil {
		fmt.Println("err:", r.Err)
		return
	}

	fmt.Println("fetched", c2.Fetch("Example").Url)

	fmt.Println("====")

	fmt.Println("begin dial")

	c := client.Dial("/", &client.Auth{Name: "n", Token: "t"})
	if c.Err != nil {
		fmt.Println("err:", c.Err)
		return
	}

	fmt.Println("dialed", c)

	var Output []int

	reply := c.Invoke("Example:rpc", []int{1, 2, 3}).Scan(&Output)

	if reply.Err != nil {
		fmt.Println("err:", reply.Err)
		return
	}

	fmt.Println("reply", reply)
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

	fmt.Println("====")

	obj := &wire.Example

	b, err := json.Marshal(obj)

	if err != nil {
		fmt.Println("err", err)
	} else {
		fmt.Println("json", string(b))
	}

	var variant wire.Variant

	err = json.Unmarshal(b, &variant)

	if err != nil {
		fmt.Println("err", err)
	} else {
		fmt.Println("variant contents", variant.Kind, variant.Msg)
	}

	var inst wire.Instance

	variant.Unwrap(&inst)

	fmt.Println("Unwrap", inst)
}
