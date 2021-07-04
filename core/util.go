package core

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

type ParamSet map[string]string
type Method string

var (
	Client *http.Client
)

func DoRequest(targeturl string, method Method, params ParamSet) (*http.Response, error) {
	if method == http.MethodGet {
		return reqGet(targeturl, params)
	}

	if method == http.MethodPost {
		return reqPost(targeturl, params)
	}
	return nil, errors.New("unsupported method")
}

func reqPost(targeturl string, params ParamSet) (*http.Response, error) {
	data := url.Values{}
	for k := range params {
		data.Set(k, params[k])
	}

	u, _ := url.ParseRequestURI(targeturl)
	urlStr := u.String()
	// client := &http.Client{}
	r, _ := http.NewRequest(http.MethodPost, urlStr, strings.NewReader(data.Encode())) // URL-encoded payload
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	r.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))

	resp, err := Client.Do(r)
	if err != nil {

		return nil, err
	}
	// defer resp.Body.Close()
	return resp, nil
}

func reqGet(targeturl string, params ParamSet) (*http.Response, error) {
	u, _ := url.ParseRequestURI(targeturl)
	urlStr := u.String()
	r, _ := http.NewRequest(http.MethodGet, urlStr, nil) // URL-encoded payload

	q := r.URL.Query()

	for k := range params {
		q.Add(k, params[k])
	}

	r.URL.RawQuery = q.Encode()

	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	// r.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))

	resp, err := Client.Do(r)
	if err != nil {
		return nil, err
	}
	// defer resp.Body.Close()
	return resp, nil
}

//randomString returns a random string of length of parameter n
func RandomString(n int) string {
	rand.Seed(time.Now().UnixNano())
	var chars = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	s := make([]rune, n)
	for i := range s {
		s[i] = chars[rand.Intn(len(chars))]
	}
	return string(s)
}

//randomString returns a random string of length of parameter n
func RandomNumString(n int) string {
	rand.Seed(time.Now().UnixNano())
	var chars = []rune("0123456789")
	s := make([]rune, n)
	for i := range s {
		s[i] = chars[rand.Intn(len(chars))]
	}
	return string(s)
}

func GetMD5Hash(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}

// readLines reads a whole file into memory and returns a slice of its lines.
func ReadLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

// GetBodyString returns a string or empty from a *http.response, and resets resp.Body to its original state so it can be read again.
func GetBodyString(resp *http.Response) string {
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("[ERROR] failed to read response body")
		return ""
	}
	resp.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
	bodyString := string(bodyBytes)
	return bodyString
}
