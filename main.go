package main

import (
	"fmt"
	"os"

	"github.com/driftsec/parabuster/core"
	"github.com/driftsec/parabuster/find"

	"github.com/driftsec/parabuster/attack"
)

// func init() {
// 	cfg := &tls.Config{
// 		InsecureSkipVerify: true,
// 	}

// 	http.DefaultClient.Transport = &http.Transport{
// 		TLSClientConfig: cfg,
// 	}

// 	tr := &http.Transport{
// 		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
// 	}
// 	cookieJar, _ := cookiejar.New(nil)
// 	core.Client = &http.Client{Transport: tr,
// 		Jar: cookieJar,
// 		CheckRedirect: func(req *http.Request, via []*http.Request) error {
// 			return http.ErrUseLastResponse
// 		}}

// }

func main() {
	header()
	if len(os.Args) < 2 {
		core.Eprint("expected 'find' or 'attack' subcommands")
		fmt.Println("")
		paraUsage()
		os.Exit(1)
	}

	switch os.Args[1] {

	case "find":
		find.Flags.Parse(os.Args[2:])
		find.FindMain()

	case "attack":
		attack.Flags.Parse(os.Args[2:])
		attack.AttackMain()

	default:
		core.Eprint("expected 'find' or 'attack' subcommands")
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
	fmt.Println()

}

func header() {
	header := `
	
__________                    __________                __                
\______   \_____ ____________ \______   \__ __  _______/  |_  ___________ 
 |     ___/\__  \\_  __ \__  \ |    |  _/  |  \/  ___/\   __\/ __ \_  __ \
 |    |     / __ \|  | \// __ \|    |   \  |  /\___ \  |  | \  ___/|  | \/
 |____|    (____  /__|  (____  /______  /____//____  > |__|  \___  >__|   
                \/           \/       \/           \/            \/       
`
	fmt.Println(core.ErrorColor, header, core.Reset)
}
