package attack

import (
	"flag"
	"fmt"
)

var Flags = flag.NewFlagSet("attack", flag.ExitOnError)
var URL = Flags.String("url", "", "Target URL to test")
var Payloads = Flags.String("payloads", "", "File containing payloads")

func Usage() {
	fmt.Println()
	Flags.Usage()
}

func AttackMain() {
	fmt.Println("a")
}
