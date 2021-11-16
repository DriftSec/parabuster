package core

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type RequestConfig struct {
	Method      string
	ContentType string
	Headers     map[string]string
	Query       map[string]string
	Params      map[string]string
	MultiPart   []map[string]string
	Url         string
	RawBody     string
	Scheme      string
}

var TemplateXML = `<?xml version="1.0" ?><root>BUSTHERE</root>`

type ParamSet map[string]string
type HeaderSet map[string]string
type Method string

var (
	Client *http.Client
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
	Client = &http.Client{Transport: tr,
		Jar: cookieJar,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}}

}

func DoRequest(targeturl string, methodChar string, params ParamSet, extraHeaders HeaderSet) (*http.Response, error) {

	if methodChar == "g" {
		req, _ := CreateReqGet(targeturl, params, extraHeaders)
		return MakeRequest(req, "")
	}
	if methodChar == "p" {
		req, _ := CreateReqPost(targeturl, params, extraHeaders)
		return MakeRequest(req, "")
	}
	if methodChar == "m" {
		req, _ := CreateReqPostMultipart(targeturl, params, extraHeaders)
		return MakeRequest(req, "")
	}
	if methodChar == "x" {
		req, _ := CreateReqPostXML(targeturl, params, extraHeaders)
		return MakeRequest(req, "")
	}
	if methodChar == "j" {
		req, _ := CreateReqPostJSON(targeturl, params, extraHeaders)
		return MakeRequest(req, "")
	}

	return nil, errors.New("unsupported method")
}

func CreateReqPostXML(targeturl string, params ParamSet, extraHeaders HeaderSet) (*http.Request, error) {
	data := strings.Replace(TemplateXML, "BUSTHERE", params2XML(params), -1)
	u, _ := url.ParseRequestURI(targeturl)
	urlStr := u.String()
	// client := &http.Client{}
	r, _ := http.NewRequest(http.MethodPost, urlStr, bytes.NewBuffer([]byte(data))) // URL-encoded payload
	for k, v := range extraHeaders {
		r.Header.Add(k, v)
	}
	r.Header.Add("Content-Type", "application/xml; charset=utf-8")
	r.Header.Add("Content-Length", strconv.Itoa(len(data)))

	return r, nil
}

// StringMap marshals into XML.
func params2XML(p ParamSet) string {
	var ret string
	for k, v := range p {
		start := "<" + k + ">"
		end := "</" + k + ">"
		ret += start + v + end
	}
	return ret
}

func CreateReqPostJSON(targeturl string, params ParamSet, extraHeaders HeaderSet) (*http.Request, error) {
	data, err := json.Marshal(&params)
	if err != nil {
		fmt.Println(err)
	}

	u, _ := url.ParseRequestURI(targeturl)
	urlStr := u.String()
	// client := &http.Client{}
	r, _ := http.NewRequest(http.MethodPost, urlStr, bytes.NewBuffer([]byte(data))) // URL-encoded payload
	for k, v := range extraHeaders {
		r.Header.Add(k, v)
	}
	r.Header.Add("Content-Type", "application/json")
	r.Header.Add("Content-Length", strconv.Itoa(len(data)))

	return r, nil
}
func CreateReqPostMultipart(targeturl string, params ParamSet, extraHeaders HeaderSet) (*http.Request, error) {
	data := url.Values{}
	for k := range params {
		data.Set(k, params[k])
	}
	// ! >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>> create multipart body here

	u, _ := url.ParseRequestURI(targeturl)
	urlStr := u.String()
	// client := &http.Client{}
	r, _ := http.NewRequest(http.MethodPost, urlStr, strings.NewReader(data.Encode())) // URL-encoded payload
	for k, v := range extraHeaders {
		r.Header.Add(k, v)
	}
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	r.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))

	return r, nil
}

func CreateReqPost(targeturl string, params ParamSet, extraHeaders HeaderSet) (*http.Request, error) {
	data := url.Values{}
	for k := range params {
		data.Set(k, params[k])
	}

	u, _ := url.ParseRequestURI(targeturl)
	urlStr := u.String()
	// client := &http.Client{}
	r, _ := http.NewRequest(http.MethodPost, urlStr, strings.NewReader(data.Encode())) // URL-encoded payload
	for k, v := range extraHeaders {
		r.Header.Add(k, v)
	}
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	r.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))

	return r, nil
}

func CreateReqGet(targeturl string, params ParamSet, extraHeaders HeaderSet) (*http.Request, error) {
	u, _ := url.ParseRequestURI(targeturl)
	urlStr := u.String()
	r, _ := http.NewRequest(http.MethodGet, urlStr, nil) // URL-encoded payload

	q := r.URL.Query()

	for k := range params {
		q.Add(k, params[k])
	}

	r.URL.RawQuery = q.Encode()
	for k, v := range extraHeaders {
		r.Header.Add(k, v)
	}
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	return r, nil
}

func MakeRequest(r *http.Request, proxy string) (*http.Response, error) {
	if proxy != "" {
		proxyUrl, _ := url.Parse(proxy)
		Client.Transport = &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, Proxy: http.ProxyURL(proxyUrl)}
	}

	resp, err := Client.Do(r)
	if err != nil {
		return nil, err
	}
	// defer resp.Body.Close()
	Dprint(r.URL, resp.StatusCode)
	return resp, nil
}

func ParseBodyXML(conf *RequestConfig) {

}

func ParseBodyJSON(conf *RequestConfig) {

}

func ParseBodyMultiPart(conf *RequestConfig) error {
	//multipart/form-data; boundary=------------------------948a6137eef50079
	tmp := strings.TrimSpace(conf.ContentType)
	parts := strings.Split(tmp, ";")
	Sections := []map[string]string{}
	if len(parts) != 2 {
		return errors.New("failed to parse form boundry")
	}
	boundry := strings.ReplaceAll(strings.Split(parts[1], "=")[1], "-", "")

	tmpBody := conf.RawBody

	end := regexp.MustCompile("[-]+" + boundry + "--$")
	tmpBody = end.ReplaceAllString(tmpBody, "")

	re := regexp.MustCompile("[-]+" + boundry + "\n")
	split := re.Split(tmpBody, -1)

	for _, i := range split {
		if len(i) == 0 {
			continue
		}
		Sections = append(Sections, mapBoundry(i))
	}
	conf.MultiPart = Sections
	return nil

}

func mapBoundry(section string) map[string]string {
	ret := make(map[string]string)
	// data := []string{}
	parts := strings.SplitN(section, "\n\n", 2)

	header := parts[0]
	body := parts[1]
	ret["body"] = body

	scanner := bufio.NewScanner(strings.NewReader(header))
	for scanner.Scan() {
		line := scanner.Text()
		pv := strings.SplitN(line, ":", 2)
		if strings.Contains(pv[1], "; ") {
			tmp := strings.Split(pv[1], "; ")
			ret[pv[0]] = tmp[0]
			for _, f := range tmp[1:] {
				fpv := strings.SplitN(f, "=", 2)
				ret[strings.TrimLeft(fpv[0], " ")] = fpv[1]
			}
		} else {
			ret[strings.TrimLeft(pv[0], " ")] = pv[1]
		}

	}

	return ret
}

// GetBodyString returns a string or empty from a *http.response, and resets resp.Body to its original state so it can be read again.
// func GetBodyString(resp *http.Response) string {
// 	bodyBytes, err := ioutil.ReadAll(resp.Body)
// 	if err != nil {
// 		fmt.Println("[ERROR] failed to read response body")
// 		return ""
// 	}
// 	resp.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
// 	bodyString := string(bodyBytes)
// 	return bodyString
// }

func GetReqBody(r *http.Request) io.ReadCloser   { return r.Body }
func GetRespBody(r *http.Response) io.ReadCloser { return r.Body }
func GetBodyString(rc io.ReadCloser) string {
	// defer rc.Close()
	bodyBytes, err := ioutil.ReadAll(rc)
	if err != nil {
		fmt.Println("[ERROR] failed to read response body")
		return ""
	}
	rc = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
	return string(bodyBytes)
}

func DumpRawRequest(req *http.Request, path string) {
	f, err := os.Create(path)
	if err != nil {
		fmt.Println(err)
		return
	}
	w := bufio.NewWriter(f)

	qp := ""
	if len(req.URL.RawQuery) > 0 {
		qp = "?" + req.URL.RawQuery
	}
	w.WriteString(req.Method + " " + req.URL.Path + qp + " " + req.Proto + "\n")
	w.WriteString("Host: " + req.Host + "\n")
	for k := range req.Header {
		w.WriteString(k + ": " + req.Header.Get(k) + "\n")
	}

	if req.Method != "GET" {
		w.WriteString("\n\n")
		w.WriteString(GetBodyString(req.Body))
	}
	w.WriteString("\n\n")
	w.Flush()
}

func ReadRawRequest(path string, scheme string) (*RequestConfig, error) {
	file, err := os.Open(path)
	if err != nil {
		return &RequestConfig{}, fmt.Errorf("could not open request file: %s", err)
	}
	defer file.Close()

	r := bufio.NewReader(file)

	s, err := r.ReadString('\n')
	if err != nil {
		return &RequestConfig{}, fmt.Errorf("could not read request: %s", err)
	}
	parts := strings.Split(s, " ")
	if len(parts) < 3 {
		return &RequestConfig{}, fmt.Errorf("malformed request supplied")
	}

	var conf RequestConfig
	conf.Scheme = scheme
	conf.Headers = make(map[string]string)
	conf.Params = make(map[string]string)
	conf.Query = make(map[string]string)

	// Set the request Method
	conf.Method = parts[0]

	for {
		line, err := r.ReadString('\n')
		line = strings.TrimSpace(line)

		if err != nil || line == "" {
			break
		}

		p := strings.SplitN(line, ":", 2)
		if len(p) != 2 {
			continue
		}

		if strings.EqualFold(p[0], "content-length") {
			continue
		}
		if strings.EqualFold(p[0], "content-type") {
			conf.ContentType = strings.TrimSpace(p[1])
		}

		conf.Headers[strings.TrimSpace(p[0])] = strings.TrimSpace(p[1])
	}

	// Handle case with the full http url in path. In that case,
	// ignore any host header that we encounter and use the path as request URL
	var tmpUrl string
	if strings.HasPrefix(parts[1], "http") {
		parsed, err := url.Parse(parts[1])
		if err != nil {
			return &RequestConfig{}, fmt.Errorf("could not parse request URL: %s", err)
		}

		tmpUrl = parts[1]
		conf.Headers["Host"] = parsed.Host
	} else {
		tmpUrl = conf.Scheme + "://" + conf.Headers["Host"] + parts[1]
	}

	uq := strings.SplitN(tmpUrl, "?", 2)
	conf.Url = uq[0]
	if len(uq) >= 2 {
		for _, set := range strings.Split(uq[1], "&") {
			pv := strings.Split(set, "=")
			conf.Query[pv[0]] = pv[1]
		}
	}

	// Set the request body
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return &RequestConfig{}, fmt.Errorf("could not read request body: %s", err)
	}
	conf.RawBody = string(b)

	// Remove newline (typically added by the editor) at the end of the file
	if strings.HasSuffix(conf.RawBody, "\r\n") {
		conf.RawBody = conf.RawBody[:len(conf.RawBody)-2]
	} else if strings.HasSuffix(conf.RawBody, "\n") {
		conf.RawBody = conf.RawBody[:len(conf.RawBody)-1]
	}

	if conf.ContentType == "application/x-www-form-urlencoded" {
		for _, set := range strings.Split(conf.RawBody, "&") {
			pv := strings.Split(set, "=")
			if len(pv) < 2 {
				continue
			}
			conf.Params[pv[0]] = pv[1]
		}
	}

	// ! need to deal with other content types, i.e. JSON, XML
	if conf.ContentType == "application/xml" {
		ParseBodyXML(&conf)
	}

	if conf.ContentType == "application/json" {
		ParseBodyJSON(&conf)
	}

	if strings.HasPrefix(conf.ContentType, "multipart/form-data") {
		err := ParseBodyMultiPart(&conf)
		if err != nil {
			return &RequestConfig{}, err
		}

	}

	return &conf, nil
}
