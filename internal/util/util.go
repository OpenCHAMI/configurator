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

// Wrapper function to simplify checking if a path exists.
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

// Wrapper function to simplify checking if a path is a directory.
func IsDirectory(path string) (bool, error) {
	// This returns an *os.FileInfo type
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false, fmt.Errorf("failed to stat path: %v", err)
	}

	// IsDir is short for fileInfo.Mode().IsDir()
	return fileInfo.IsDir(), nil
}

// Wrapper function to confine making a HTTP request into a single function
// instead of multiple.
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

// Returns the git commit string by executing command.
// NOTE: This currently requires git to be installed.
// TODO: Change how this is done to not require executing a command.
func GitCommit() string {
	c := exec.Command("git", "rev-parse", "HEAD")
	stdout, err := c.Output()
	if err != nil {
		return ""
	}

	return strings.TrimRight(string(stdout), "\n")
}

// General function to remove element by a given index.
// NOTE: would it be better to use slices.DeleteFunc instead?
func RemoveIndex[T comparable](s []T, index int) []T {
	ret := make([]T, 0)
	ret = append(ret, s[:index]...)
	return append(ret, s[index+1:]...)
}

// General function to copy elements from slice if condition is true.
func CopyIf[T comparable](s []T, condition func(t T) bool) []T {
	var f = make([]T, 0)
	for _, e := range s {
		if condition(e) {
			f = append(f, e)
		}
	}
	return f
}
