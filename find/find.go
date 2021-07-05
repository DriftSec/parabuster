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
var URL = Flags.String("url", "", "Target URL to test")
var Wordlist = Flags.String("wordlist", "", "Parameter wordlist")
var Chunks = Flags.Int("chunk", 50, "Chunk Size")
var Meth = Flags.String("method", "all", "Method [get,post,all]")
var MaxConcurrent = Flags.Int("threads", 10, "Concurent threads")

var FoundGet []string
var FoundPost []string

var throttle core.Throttle

func Usage() {
	fmt.Println()
	Flags.Usage()
}

func FindMain() {
	throttle = throttle.New(*MaxConcurrent)

	words, err := core.ReadLines(*Wordlist)
	if err != nil {
		core.Eprint("Failed to open the wordlist")
		os.Exit(1)
	}

	core.Iprint("Testing connection")
	resp, err := core.DoRequest(*URL, http.MethodGet, core.ParamSet{})
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

	switch strings.ToLower(*Meth) {
	case "get":
		ScanGet(words)
		fmt.Println("\033[u\033[K\n")
		if len(FoundGet) > 0 {
			core.Nprint("Found", len(FoundGet), "GET parameters:", strings.Join(FoundGet, ", "), "\n")
			fmt.Println("(GET):" + *URL + "?" + strings.Join(FoundGet, "=FUZZ&") + "=FUZZ")
			fmt.Println()
		} else {
			core.Fprint("No GET parameters found !!")
		}
	case "post":
		ScanPost(words)
	default:
		ScanGet(words)
		ScanPost(words)
	}

	fmt.Println("\033[u\033[K\n")
	if strings.ToLower(*Meth) == "get" || strings.ToLower(*Meth) == "all" {
		if len(FoundGet) > 0 {
			core.Nprint("Found", len(FoundGet), "GET parameters:", strings.Join(FoundGet, ", "), "\n")
			fmt.Println("(GET):" + *URL + "?" + strings.Join(FoundGet, "=FUZZ&") + "=FUZZ")
			fmt.Println()
		} else {
			core.Fprint("No GET parameters found !!")
		}
	}
	if strings.ToLower(*Meth) == "post" || strings.ToLower(*Meth) == "all" {
		if len(FoundPost) > 0 {
			core.Nprint("Found", len(FoundPost), "POST parameters:", strings.Join(FoundPost, ", "), "\n")
			fmt.Println("(POST):" + *URL + "?" + strings.Join(FoundPost, "=FUZZ&") + "=FUZZ")
			fmt.Println()
		} else {
			core.Fprint("No POST parameters found !!")
		}
	}
}

func ScanPost(words []string) {
	// fmt.Println("\033[u\033[K\n")
	core.Iprint("Starting Auto Calibration (POST)")
	ac, err := AutoCalibrated(*URL, http.MethodPost)
	if err != nil {
		core.Eprint("AutoCalibration Failed:", err.Error())
		os.Exit(1)
	}
	core.Iprint("Content is stable")
	core.Iprint("Running (POST)...")
	fmt.Print("\033[s")

	chunks := GetChunks(words, *Chunks)
	for i, chunk := range chunks {
		fmt.Print("\033[u\033[K", "processing chunk ", i, " of ", len(chunks))

		throttle.WaitForSpot()
		go threadFunc(*URL, http.MethodPost, ac, chunk)

	}
	throttle.WaitForDone()

}

func ScanGet(words []string) {
	core.Iprint("Starting Auto Calibration (GET)")
	ac, err := AutoCalibrated(*URL, http.MethodGet)
	if err != nil {
		core.Eprint("AutoCalibration Failed:", err.Error())
		os.Exit(1)
	}
	core.Iprint("Content is stable")
	core.Iprint("Running (GET)...")
	fmt.Print("\033[s")

	chunks := GetChunks(words, *Chunks)
	for i, chunk := range chunks { //! maybe we can make this a chan, pipe new form names to the chan and iterate??????????????
		fmt.Print("\033[u\033[K", "processing chunk ", i, " of ", len(chunks))
		throttle.WaitForSpot()
		go threadFunc(*URL, http.MethodGet, ac, chunk)

	}

	throttle.WaitForDone()

}

func threadFunc(url string, method core.Method, cal *Calibration, chunk []string) {
	defer throttle.Done()
	p := make(core.ParamSet)
	for _, a := range chunk {
		p[a] = core.RandomString(8) //! TODO try other value types ???
	}
	isdiff, msg := requestAndDiff(*URL, method, p, cal)
	if !isdiff && msg != "" {
		core.Eprint(msg)
	} else if isdiff {
		NarrowHits(*URL, method, p, cal)
	}

}

// NarrowHits recursively splits, requests and compares any hits until parameter length is 1.
func NarrowHits(url string, method core.Method, params core.ParamSet, cal *Calibration) {
	a, b := splitMap(params)
	isdiffa, reason := requestAndDiff(url, method, a, cal)
	if isdiffa {
		if len(a) == 1 {
			parseFinal(a, method, reason)
		} else {
			NarrowHits(url, method, a, cal)
		}
	}
	isdiffb, reason := requestAndDiff(url, method, b, cal)
	if isdiffb {
		if len(b) == 1 {
			parseFinal(b, method, reason)
		} else {
			NarrowHits(url, method, b, cal)
		}

	}
}

func parseFinal(p core.ParamSet, method core.Method, msg string) {
	for k := range p {
		if method == http.MethodGet {
			fmt.Printf("\033[u\033[K")
			core.Sprint("Found Parameter:", k, "(GET) ("+msg+")")
		}
		if method == http.MethodPost {
			fmt.Printf("\033[u\033[K")
			core.Sprint("Found Parameter:", k, "(POST) ("+msg+")")
		}
		fmt.Print("\033[s")
		if method == http.MethodGet {
			FoundGet = append(FoundGet, k)
		}
		if method == http.MethodPost {
			FoundPost = append(FoundPost, k)
		}
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
func requestAndDiff(url string, method core.Method, params core.ParamSet, cal *Calibration) (bool, string) {
	resp, err := core.DoRequest(url, method, params)
	if err != nil {
		return false, err.Error()
	}
	newBody := core.GetBodyString(resp)
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
