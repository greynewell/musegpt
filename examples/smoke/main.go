package main

import (
	"context"
	"fmt"

	"github.com/greynewell/infermux/grpcclient"
)

func main() {
	c, err := grpcclient.New("127.0.0.1:8601")
	if err != nil {
		panic(err)
	}
	defer c.Close()
	res, err := c.Prompt(context.Background(), "echo-v1", "smoke test")
	if err != nil {
		panic(err)
	}
	fmt.Printf("content=%q provider=%s tokens_in=%d tokens_out=%d cost=%.6f\n",
		res.Content, res.Provider, res.TokensIn, res.TokensOut, res.CostUSD)
	ps, err := c.ListProviders(context.Background())
	if err != nil {
		panic(err)
	}
	fmt.Printf("providers=%+v\n", ps)
}
