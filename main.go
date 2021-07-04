package main

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"os"
	"parabuster/attack"
	"parabuster/core"
	"parabuster/find"
)

func init() {
	cfg := &tls.Config{
		InsecureSkipVerify: true,
	}

	http.DefaultClient.Transport = &http.Transport{
		TLSClientConfig: cfg,
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	cookieJar, _ := cookiejar.New(nil)
	core.Client = &http.Client{Transport: tr,
		Jar: cookieJar,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}}

}

func main() {
	header()
	if len(os.Args) < 2 {
		fmt.Println("expected 'find' or 'attack' subcommands")
		fmt.Println("")
		paraUsage()
		os.Exit(1)
	}

	switch os.Args[1] {

	case "find":
		if len(os.Args[2:]) < 2 {
			find.Flags.Usage()
		}
		find.Flags.Parse(os.Args[2:])
		find.FindMain()

	case "attack":
		attack.Flags.Parse(os.Args[2:])
		if len(os.Args[2:]) < 2 {
			attack.Flags.Usage()
		}

		attack.AttackMain()

	default:
		fmt.Println("expected 'find' or 'attack' subcommands")
		paraUsage()
		os.Exit(1)
	}
}

func paraUsage() {
	fmt.Println("Usage: parabuster [mode] [options]")
	fmt.Println("")
	fmt.Println("Modes:")
	fmt.Println("     find                 Discovers paramaters for an URL.")
	fmt.Println("     attack               Fuzzes known parameters for issues.")
	find.Usage()
	attack.Usage()

}

func header() {
	header := `
__________                            __________                __                
\______   \_____ ____________    _____\______   \__ __  _______/  |_  ___________ 
 |     ___/\__  \\_  __ \__  \  /     \|    |  _/  |  \/  ___/\   __\/ __ \_  __ \
 |    |     / __ \|  | \// __ \|  Y Y  \    |   \  |  /\___ \  |  | \  ___/|  | \/
 |____|    (____  /__|  (____  /__|_|  /______  /____//____  > |__|  \___  >__|   
                \/           \/      \/       \/           \/            \/       
                                                                             
`
	fmt.Println(header)
}
