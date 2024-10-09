package util

import (
	"archive/tar"
	"bytes"
	"cmp"
	"compress/gzip"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"slices"
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
		return false, fmt.Errorf("failed to stat path (%s): %v", path, err)
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
// NOTE: This also requires running within the repository.
// TODO: Change how this is done to not require executing a command.
func GitCommit() string {
	c := exec.Command("git", "rev-parse", "--short=8", "HEAD")
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

func RemoveDuplicates[T cmp.Ordered](s []T) []T {
	slices.Sort(s)
	return slices.Compact(s)
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

func CreateArchive(files []string, buf io.Writer) error {
	// Create new Writers for gzip and tar
	// These writers are chained. Writing to the tar writer will
	// write to the gzip writer which in turn will write to
	// the "buf" writer
	gw := gzip.NewWriter(buf)
	defer gw.Close()
	tw := tar.NewWriter(gw)
	defer tw.Close()

	// Iterate over files and add them to the tar archive
	for _, file := range files {
		err := addToArchive(tw, file)
		if err != nil {
			return err
		}
	}

	return nil
}

func addToArchive(tw *tar.Writer, filename string) error {
	// open file to write to archive
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// get FileInfo for file size, mode, etc.
	info, err := file.Stat()
	if err != nil {
		return err
	}

	// create a tar Header from the FileInfo data
	header, err := tar.FileInfoHeader(info, info.Name())
	if err != nil {
		return err
	}

	// use full path as name (FileInfoHeader only takes the basename) to
	// preserve directory structure
	// see for more info: https://golang.org/src/archive/tar/common.go?#L626
	header.Name = filename

	// Write file header to the tar archive
	err = tw.WriteHeader(header)
	if err != nil {
		return err
	}

	// copy file content to tar archive
	_, err = io.Copy(tw, file)
	if err != nil {
		return err
	}

	return nil
}
