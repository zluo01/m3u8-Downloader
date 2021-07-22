package request

import (
	"errors"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type ReqClient struct {
	client *http.Client
}

func New(timeout time.Duration, proxy string) (*ReqClient, error) {
	reqClient := &ReqClient{
		client: http.DefaultClient,
	}

	if timeout > 0 {
		reqClient.client.Timeout = timeout
	}

	if proxy != "" {
		p, err := url.Parse(proxy)
		if err != nil {
			return nil, err
		}

		t := http.DefaultTransport.(*http.Transport).Clone()
		t.Proxy = func(*http.Request) (*url.URL, error) {
			return p, nil
		}
		reqClient.client.Transport = t
	}

	return reqClient, nil
}

func (r *ReqClient) Get(url string, headers map[string]string, retry int) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/83.0.4103.61 Safari/537.36")

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	var body []byte
	var code int

	var count int
	for count < retry {
		count++

		var resp *http.Response
		resp, err = r.client.Do(req)
		if err != nil {
			continue
		}

		body, err = io.ReadAll(resp.Body)
		resp.Body.Close()

		if err != nil {
			continue
		}

		code = resp.StatusCode

		break
	}

	if err != nil {
		return nil, err
	}

	if code != http.StatusOK || len(body) == 0 {
		return nil, errors.New("http code: " + strconv.Itoa(code))
	}

	return body, nil
}
