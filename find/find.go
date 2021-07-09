package find

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"parabuster/core"
	"reflect"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

var Flags = flag.NewFlagSet("find", flag.ExitOnError)

var URL string
var Wordlist string
var Chunks int
var Meth string
var MaxConcurrent int
var OutRequest bool
var OutBurp string

type HeaderSlice []string
type HeaderSet map[string]string

var ExtraHeaderSlice HeaderSlice
var ExtraHeaders HeaderSet

var FoundGet []string
var FoundPost []string
var Found []string
var throttle core.Throttle
var MethodText = map[string]string{
	"a": "ALL",
	"g": "GET",
	"p": "POST",
	"m": "MULTIPART",
	"x": "POSTXML",
	"j": "POSTJSON",
}

func (i *HeaderSlice) String() string {
	return fmt.Sprintf("%s", *i)
}

func (i *HeaderSlice) Set(value string) error {
	*i = append(*i, value)
	return nil
}

func init() {
	Flags.StringVar(&URL, "url", "", "Target URL to test")
	Flags.StringVar(&URL, "u", "", "")
	Flags.StringVar(&Wordlist, "wordlist", "", "Parameter wordlist")
	Flags.StringVar(&Wordlist, "w", "", "")
	Flags.IntVar(&Chunks, "chunk", 50, "Chunk Size")
	Flags.IntVar(&Chunks, "c", 50, "")
	Flags.StringVar(&Meth, "method", "gp", "Method [a:all, g:GET, p:POST normal, m:POST Multipart, x:POST XML, j:POST JSON]")
	Flags.StringVar(&Meth, "m", "gp", "")
	Flags.IntVar(&MaxConcurrent, "threads", 10, "")
	Flags.IntVar(&MaxConcurrent, "t", 10, "")
	Flags.BoolVar(&OutRequest, "oR", false, "")
	Flags.StringVar(&OutBurp, "burp", "", "")
	Flags.Var(&ExtraHeaderSlice, "H", "")
}

func Usage() {
	fmt.Println()
	// Flags.Usage()
	use := `
Usage of find:
	-chunk|c [int]
		Chunk Size (default 50)

	-method|m [method string]
		Method [a:all, g:GET, p:POST normal, m:POST Multipart, x:POST XML, j:POST JSON] (default "gp")

	-threads|t [int]
		Concurent threads (default 10)

	-url|u [url]
		Target URL to test
		
	-wordlist|w [path]
		Parameter wordlist

	-oR
		Output request files

	-burp [burp url]
		Pass a request containing found parameters to burp

	-H ["header:value"]
		Add additional headers, can be used multiple times.
	`
	fmt.Println(use)
}

func FindMain() {
	ExtraHeaders = make(HeaderSet)
	throttle = throttle.New(MaxConcurrent)

	words, err := core.ReadLines(Wordlist)
	if err != nil {
		core.Eprint("Failed to open the wordlist")
		os.Exit(1)
	}

	core.Iprint("Testing connection")
	resp, err := core.DoRequest(URL, "g", core.ParamSet{}, ExtraHeaders)
	if err != nil {
		core.Eprint(err.Error())
		os.Exit(1)
	}
	form := ExtractForm(resp)
	if len(form) > 0 {
		core.Sprint("Adding form values to queue:", form)
		var tmpWords []string
		tmpWords = append(tmpWords, form...)
		words = append(tmpWords, words...)
	}

	Meth = strings.ToLower(Meth)
	if strings.Contains(Meth, "g") || strings.Contains(Meth, "a") {
		Scan(words, "g")
	}
	if strings.Contains(Meth, "p") || strings.Contains(Meth, "a") {
		Scan(words, "p")
	}
	if strings.Contains(Meth, "m") || strings.Contains(Meth, "a") {
		Scan(words, "m")
	}
	if strings.Contains(Meth, "x") || strings.Contains(Meth, "a") {
		Scan(words, "x")
	}
	if strings.Contains(Meth, "j") || strings.Contains(Meth, "a") {
		Scan(words, "j")
	}

	fmt.Println("\033[u\033[K")
	if len(Found) > 0 {
		core.Nprint("Found", len(Found), "parameters:", strings.Join(Found, ", "), "\n")
		parseOutput(Found)
	} else {
		core.Fprint("No parameters found !!")
	}

}

func parseOutput(found []string) {
	results := make(map[string][]string)
	for _, hit := range found {
		parts := strings.SplitN(hit, ":", 2)
		m, p := parts[0], parts[1]
		results[m] = append(results[m], p)
	}
	for k, v := range results {
		fmt.Println("(" + k + "):" + strings.Join(v, ", "))
	}

	for k, v := range results {
		params := make(core.ParamSet)
		for _, param := range v {
			params[param] = "FUZZ"
		}
		switch k {
		case "GET":
			req, _ := core.CreateReqGet(URL, params)
			if OutBurp != "" {
				core.MakeRequest(req, OutBurp)
			}
			if OutRequest {
				core.DumpRawRequest(req, k+".req")
			}
		case "POST":
			req, _ := core.CreateReqPost(URL, params)
			if OutBurp != "" {
				core.MakeRequest(req, OutBurp)
			}
			if OutRequest {
				core.DumpRawRequest(req, k+".req")
			}
		case "POSTJSON":
			req, _ := core.CreateReqPostJSON(URL, params)
			if OutBurp != "" {
				core.MakeRequest(req, OutBurp)
			}
			if OutRequest {
				core.DumpRawRequest(req, k+".req")
			}
		case "POSTXML":
			req, _ := core.CreateReqPostXML(URL, params)
			if OutBurp != "" {
				core.MakeRequest(req, OutBurp)
			}
			if OutRequest {
				core.DumpRawRequest(req, k+".req")
			}
		case "MULTIPART":
			req, _ := core.CreateReqPostMultipart(URL, params)
			if OutBurp != "" {
				core.MakeRequest(req, OutBurp)
			}
			if OutRequest {
				core.DumpRawRequest(req, k+".req")
			}
		}
	}
}

func Scan(words []string, methodChar string) {
	core.Iprint("Starting Auto Calibration (" + MethodText[methodChar] + ")")
	ac, err := AutoCalibrate(URL, methodChar)
	if err != nil {
		fmt.Println()
		core.Eprint("AutoCalibration Failed:", err.Error())
		return
	}
	core.Iprint("Content is stable")
	core.Iprint("Running (" + MethodText[methodChar] + ")...")
	fmt.Print("\033[s")

	chunks := GetChunks(words, Chunks)
	for i, chunk := range chunks {
		fmt.Print("\033[u\033[K", "processing chunk ", i, " of ", len(chunks))
		throttle.WaitForSpot()
		go threadFunc(URL, methodChar, ac, chunk)

	}

	throttle.WaitForDone()

}

func threadFunc(url string, methodChar string, cal *Calibration, chunk []string) {
	defer throttle.Done()
	p := make(core.ParamSet)
	for _, a := range chunk {
		p[a] = core.RandomString(8) //! TODO try other value types ???
	}
	isdiff, msg := requestAndDiff(URL, methodChar, p, cal)
	if !isdiff && msg != "" {
		core.Eprint(msg)
	} else if isdiff {
		NarrowHits(URL, methodChar, p, cal)
	}

}

// NarrowHits recursively splits, requests and compares any hits until parameter length is 1.
func NarrowHits(url string, methodChar string, params core.ParamSet, cal *Calibration) {
	a, b := splitMap(params)
	isdiffa, reason := requestAndDiff(url, methodChar, a, cal)
	if isdiffa {
		if len(a) == 1 {
			parseFinal(a, methodChar, reason)
		} else {
			NarrowHits(url, methodChar, a, cal)
		}
	}
	isdiffb, reason := requestAndDiff(url, methodChar, b, cal)
	if isdiffb {
		if len(b) == 1 {
			parseFinal(b, methodChar, reason)
		} else {
			NarrowHits(url, methodChar, b, cal)
		}

	}
}

func parseFinal(p core.ParamSet, methodChar string, msg string) {
	for k := range p {
		fmt.Printf("\033[u\033[K")
		core.Sprint("Found Parameter:", k, "("+MethodText[methodChar]+") ("+msg+")")

		fmt.Print("\033[s")
		Found = append(Found, MethodText[methodChar]+":"+k)
		break
	}
}

//splitMap splits a ParamSet into 2 equal chunks.
func splitMap(params core.ParamSet) (core.ParamSet, core.ParamSet) {
	n := 1
	odds := make(core.ParamSet)
	evens := make(core.ParamSet)
	for k, v := range params {
		if n%2 == 0 {
			evens[k] = v
		} else {
			odds[k] = v
		}
		n++
	}
	return odds, evens
}

// requestAndDiff makes a request and compares it to the calibration.  returns true,"reason" if different, false,"error" if error, false,"" if the same
func requestAndDiff(url string, methodChar string, params core.ParamSet, cal *Calibration) (bool, string) {
	resp, err := core.DoRequest(url, methodChar, params)
	if err != nil {
		return false, err.Error()
	}
	newBody := core.GetBodyString(resp.Body)
	curHeaders := resp.Header
	delete(curHeaders, "Date")
	delete(curHeaders, "Content-Length")
	curStatus := resp.Status

	if curStatus != cal.Status {
		return true, "status changed"
	}

	if !reflect.DeepEqual(curHeaders, cal.Headers) {
		return true, "headers changed"
	}

	if isBodyDiff(cal.BaseBody, newBody, params) {
		return true, "body changed"
	}
	return false, ""
}

// GetChunks splits a wordlists slice into parameter chunks
func GetChunks(xs []string, chunkSize int) [][]string {
	if len(xs) == 0 {
		return nil
	}
	divided := make([][]string, (len(xs)+chunkSize-1)/chunkSize)
	prev := 0
	i := 0
	till := len(xs) - chunkSize
	for prev < till {
		next := prev + chunkSize
		divided[i] = xs[prev:next]
		prev = next
		i++
	}
	divided[i] = xs[prev:]
	return divided
}

func ExtractForm(resp *http.Response) []string {
	var params []string
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		core.Eprint("Error loading HTTP response body. ", err)
		return []string{}
	}
	doc.Find("form").Each(func(_ int, s *goquery.Selection) {
		s.Find("input").Each(func(_ int, s *goquery.Selection) {
			name, _ := s.Attr("name")
			if name == "" {
				return
			}
			params = append(params, name)

		})
		s.Find("textarea").Each(func(_ int, s *goquery.Selection) {
			name, _ := s.Attr("name")
			if name == "" {
				return
			}
			params = append(params, name)
		})
	})
	return params
}
