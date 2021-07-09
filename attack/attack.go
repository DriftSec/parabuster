package attack

import (
	"flag"
	"fmt"
	"log"
	"parabuster/core"
)

var Flags = flag.NewFlagSet("attack", flag.ExitOnError)

var URL string
var Wordlist string
var Chunks int
var Meth string
var MaxConcurrent int
var RequestFile string

var throttle core.Throttle

func init() {
	Flags.StringVar(&URL, "url", "", "Target URL to test")
	Flags.StringVar(&URL, "u", "", "")
	Flags.StringVar(&RequestFile, "request", "", "Read a request from a file")
	Flags.StringVar(&RequestFile, "r", "", "")
	Flags.StringVar(&Wordlist, "payloads", "", "Payload wordlist")
	Flags.StringVar(&Wordlist, "p", "", "")
	Flags.StringVar(&Meth, "method", "all", "Method [get,post,all]")
	Flags.StringVar(&Meth, "m", "all", "")
	Flags.IntVar(&MaxConcurrent, "threads", 10, "Concurent threads")
	Flags.IntVar(&MaxConcurrent, "t", 10, "")
}

func Usage() {
	fmt.Println()
	// Flags.Usage()
	use := `
Usage of Attack:
	-method|m string
		Method [get,post,all] (default "all")

	-threads|t int
		Concurent threads (default 10)
	
	-request|r string
	Read a request from a file

	-url|u string
		Target URL to test
		
	-payloads|p string
		Payload wordlist

Injection points:
	can be marked with "parameter=FUZZ" or burp style "parameter=§value§" 
	`
	fmt.Println(use)
}

func AttackMain() {
	throttle = throttle.New(MaxConcurrent)

	if URL == "" && RequestFile == "" {
		core.Eprint("no target provided, you specify an url or a request file!!!")
	}

	// core.RequestFromFile(RequestFile)

	tmp, err := core.ReadRawRequest(RequestFile, "http")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(tmp.ContentType)
	fmt.Println(tmp.MultiPart)

	// words, err := core.ReadLines(Wordlist)
	// if err != nil {
	// 	core.Eprint("Failed to open the payload wordlist")
	// 	os.Exit(1)
	// }

	// core.Iprint("Testing connection")
	// resp, err := core.DoRequest(URL, http.MethodGet, core.ParamSet{})
	// if err != nil {
	// 	core.Eprint(err.Error())
	// 	os.Exit(1)
	// }
	// fmt.Println(len(words), resp)
}
