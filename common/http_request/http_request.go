package http_request

import (
	"io/ioutil"
	"net/http"
)

// HttpGet Get请求
func HttpGet(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return []byte(""), err
	}
	defer resp.Body.Close()
	result, err := ioutil.ReadAll(resp.Body)
	return result, err
}
