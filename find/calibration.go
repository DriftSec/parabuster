package find

import (
	"errors"
	"net/http"
	"parabuster/core"
	"reflect"
	"regexp"
	"strings"
)

type Calibration struct {
	Baseline *http.Response
	BaseBody string
	Headers  http.Header
	Status   string
}

func AutoCalibrated(url string, method core.Method) (*Calibration, error) {
	var (
		cal  Calibration
		resp *http.Response
		err  error
	)

	// Baseline
	tmpParams := make(core.ParamSet)
	testParam := core.RandomString(5)
	testVal := core.RandomString(6)
	tmpParams[testParam] = testVal
	resp, err = core.DoRequest(url, method, tmpParams)
	if err != nil {
		return &Calibration{}, err
	}
	origBody := core.GetBodyString(resp)
	origBody = strings.ReplaceAll(origBody, testParam, "") // need to clean random strings from body for later comparisions
	origBody = strings.ReplaceAll(origBody, testVal, "")   // need to clean random strings from body for later comparisions
	cal.Baseline = resp
	cal.BaseBody = origBody
	cal.Headers = resp.Header
	delete(cal.Headers, "Date")
	delete(cal.Headers, "Content-Length")
	cal.Status = resp.Status

	// repeat 5 iterations and compare
	tmpParams2 := make(core.ParamSet)
	for i := 0; i <= 5; i++ {
		tmpParams2[core.RandomString(i+3)] = core.RandomString(i + 4)
		resp, err = core.DoRequest(url, method, tmpParams2)
		if err != nil {
			return &Calibration{}, err
		}
		newBody := core.GetBodyString(resp)

		curHeaders := resp.Header
		delete(curHeaders, "Date")
		delete(curHeaders, "Content-Length")
		curStatus := resp.Status

		if curStatus != cal.Status {
			return &Calibration{}, errors.New("content is not stable (status)")
		}

		if !reflect.DeepEqual(curHeaders, cal.Headers) {
			return &Calibration{}, errors.New("content is not stable (headers)")
		}

		if isBodyDiff(cal.BaseBody, newBody, tmpParams2) {
			return &Calibration{}, errors.New("content is not stable (body)")
		}
	}
	return &cal, nil
}

func isBodyDiff(cleanOrig string, newBody string, params core.ParamSet) bool {
	orig := splitBody(cleanOrig)
	orig = cleanReflection(orig, params)
	new := splitBody(newBody)
	new = cleanReflection(new, params)
	return !reflect.DeepEqual(orig, new)
}

func cleanReflection(body []string, params core.ParamSet) []string {
	body = Unique(body)
	for k := range params {
		body = remove(body, k)
		body = remove(body, params[k])
	}
	return body
}

func remove(l []string, item string) []string {
	for i, other := range l {
		if other == item {
			l = append(l[:i], l[i+1:]...)
		}
	}
	return l
}

func splitBody(body string) []string {
	return regexp.MustCompile("[\\:\\,\\.\\s\\<\\>\\/\\=\"\\']+").Split(body, -1)
}

func Unique(strSlice []string) []string {
	keys := make(map[string]bool)
	list := []string{}
	for _, entry := range strSlice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

func Contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
