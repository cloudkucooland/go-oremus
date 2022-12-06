package main

import (
	"context"
	"fmt"
	"os"

	"github.com/cloudkucooland/go-oremus"
)

func main() {
	ctx := context.Background()

	for i := 1; i < len(os.Args); i++ {
        ref, err := oremus.CleanReference(os.Args[i])
		if err != nil {
			panic(err)
		}

		res, err := oremus.Get(ctx, ref)
		if err != nil {
			panic(err)
		}
		fmt.Println(res)
	}
}
