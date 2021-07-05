package find

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"parabuster/core"
	"reflect"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
)

var Flags = flag.NewFlagSet("find", flag.ExitOnError)
var URL = Flags.String("url", "", "Target URL to test")
var Wordlist = Flags.String("wordlist", "", "Parameter wordlist")
var Chunks = Flags.Int("chunk", 50, "Chunk Size")
var Threads = Flags.Int("threads", 10, "Concurent threads")

type Threader struct {
	wC  int
	wG  sync.WaitGroup
	mux sync.Mutex
}

var threads Threader
var FoundGet []string
var FoundPost []string

func Usage() {
	fmt.Println()
	Flags.Usage()
}

func FindMain() {
	words, err := core.ReadLines(*Wordlist)
	if err != nil {
		fmt.Println("[ERROR] Failed to open the wordlist")
		os.Exit(1)
	}

	fmt.Println("[!] Testing connection")
	resp, err := core.DoRequest(*URL, http.MethodGet, core.ParamSet{})
	if err != nil {
		fmt.Println("[ERROR]", err.Error())
	}
	form := ExtractForm(resp)
	if len(form) > 0 {
		fmt.Println("[+] Adding form values to queue:", form)
		var tmpWords []string
		tmpWords = append(tmpWords, form...)
		words = append(tmpWords, words...)
	}

	ScanGet(words)
	ScanPost(words)

	fmt.Println("\033[u\033[K\n")
	if len(FoundGet) > 0 {
		fmt.Println("[+] Found", len(FoundGet), "GET parameters:", strings.Join(FoundGet, ", "))

	} else {
		fmt.Println("[-] No GET parameters found !!")
	}

	if len(FoundPost) > 0 {
		fmt.Println("[+] Found", len(FoundPost), "POST parameters:", strings.Join(FoundPost, ", "))
		fmt.Println()
		tmp := *URL
		tmp += "?"
		for _, i := range FoundGet {
			tmp += i + "=FUZZ&"
		}
		tmp = strings.TrimRight(tmp, "&")
		fmt.Println(tmp)
	} else {
		fmt.Println("[-] No POST parameters found !!")
	}
}

func ScanPost(words []string) {
	fmt.Println("\033[u\033[K\n[!] Starting Auto Calibration (POST)")
	ac, err := AutoCalibrated(*URL, http.MethodPost)
	if err != nil {
		fmt.Println("[ERROR] AutoCalibration Failed:", err.Error())
		os.Exit(1)
	}
	fmt.Println("[+] Content is stable")
	fmt.Println("[!] Running (POST)...")
	fmt.Println()
	fmt.Print("\033[s")

	chunks := GetChunks(words, *Chunks)
	for i, chunk := range chunks {
		fmt.Print("\033[u\033[K", "processing chunk ", i, " of ", len(chunks))
		if threads.wC < *Threads {
			threads.wC++
			threads.wG.Add(1)
			go threadFunc(*URL, http.MethodPost, ac, chunk)
		} else {
			threads.wG.Wait()
		}
	}
	threads.wG.Wait()

}

func ScanGet(words []string) {
	fmt.Println("[!] Starting Auto Calibration (GET)")
	ac, err := AutoCalibrated(*URL, http.MethodGet)
	if err != nil {
		fmt.Println("[ERROR] AutoCalibration Failed:", err.Error())
		os.Exit(1)
	}
	fmt.Println("[+] Content is stable")
	fmt.Println("[!] Running (GET)...")
	fmt.Println()
	fmt.Print("\033[s")

	chunks := GetChunks(words, *Chunks)
	for i, chunk := range chunks { //! maybe we can make this a chan, pipe new form names to the chan and iterate??????????????
		fmt.Print("\033[u\033[K", "processing chunk ", i, " of ", len(chunks))
		if threads.wC < *Threads {
			threads.wC++
			threads.wG.Add(1)

			go threadFunc(*URL, http.MethodGet, ac, chunk)
		} else {
			threads.wG.Wait()
		}
		// threads.wG.Wait()

	}
	threads.wG.Wait()

}

func threadFunc(url string, method core.Method, cal *Calibration, chunk []string) {
	p := make(core.ParamSet)
	for _, a := range chunk {
		p[a] = core.RandomString(8) //! TODO try other value types ???
	}
	isdiff, msg := requestAndDiff(*URL, method, p, cal)
	if !isdiff && msg != "" {
		fmt.Println("[ERROR]", msg)
	} else if isdiff {
		NarrowHits(*URL, method, p, cal)
	}
	threads.wC--
	threads.wG.Done()
}

// NarrowHits recursively splits, requests and compares any hits until parameter length is 1.
func NarrowHits(url string, method core.Method, params core.ParamSet, cal *Calibration) {
	a, b := splitMap(params)
	isdiffa, _ := requestAndDiff(url, method, a, cal)
	if isdiffa {
		if len(a) == 1 {
			parseFinal(a, method)
		} else {
			NarrowHits(url, method, a, cal)
		}
	}
	isdiffb, _ := requestAndDiff(url, method, b, cal)
	if isdiffb {
		if len(b) == 1 {
			parseFinal(b, method)
		} else {
			NarrowHits(url, method, b, cal)
		}

	}
}

func parseFinal(p core.ParamSet, method core.Method) {
	for k := range p {
		if method == http.MethodGet {
			fmt.Println("\033[u\033[K", "[+] Found Parameter:", k, "(GET)")
		}
		if method == http.MethodPost {
			fmt.Println("\033[u\033[K", "[+] Found Parameter:", k, "(POST)")
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
		log.Fatal("Error loading HTTP response body. ", err)
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
