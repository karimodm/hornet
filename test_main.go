package main

import (
	"fmt"

	"github.com/gohornet/hornet/plugins/gossip"
)

func main() {

	fmt.Println("Trying to intialize...")
	b := []byte{1, 2, 3}
	gossip.Fuzz(b)
	fmt.Println("Done!")

}
