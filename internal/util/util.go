package util

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func MakeRequest(url string, httpMethod string, body []byte, headers map[string]string) (*http.Response, []byte, error) {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	req, err := http.NewRequest(httpMethod, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, nil, fmt.Errorf("could not create new HTTP request: %v", err)
	}
	req.Header.Add("User-Agent", "configurator")
	for k, v := range headers {
		req.Header.Add(k, v)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("could not make request: %v", err)
	}
	b, err := io.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return nil, nil, fmt.Errorf("could not read response body: %v", err)
	}
	return res, b, err
}

func ConvertMapOutput(m map[string][]byte) map[string]string {
	n := make(map[string]string, len(m))
	for k, v := range m {
		n[k] = string(v)
	}
	return n
}

func GitCommit() string {
	c := exec.Command("git", "rev-parse", "HEAD")
	stdout, err := c.Output()
	if err != nil {
		return ""
	}

	return strings.TrimRight(string(stdout), "\n")
}

// NOTE: would it be better to use slices.DeleteFunc instead
func RemoveIndex[T comparable](s []T, index int) []T {
	ret := make([]T, 0)
	ret = append(ret, s[:index]...)
	return append(ret, s[index+1:]...)
}

func CopyIf[T comparable](s []T, condition func(t T) bool) []T {
	var f = make([]T, 0)
	for _, e := range s {
		if condition(e) {
			f = append(f, e)
		}
	}
	return f
}
