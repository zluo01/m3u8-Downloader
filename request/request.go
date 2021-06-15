package request

import (
	"errors"
	"github.com/valyala/fasthttp/fasthttpproxy"
	"strconv"
	"time"

	"github.com/valyala/fasthttp"
)

type ReqClient struct {
	client *fasthttp.Client
}

func New(timeout time.Duration, proxy string) (*ReqClient, error) {
	reqClient := &ReqClient{
		client: &fasthttp.Client{
			ReadTimeout:  timeout,
			WriteTimeout: timeout,
		},
	}

	if proxy != "" {
		reqClient.client.Dial = fasthttpproxy.FasthttpHTTPDialer(proxy)
	}

	return reqClient, nil
}

func (r *ReqClient) Get(url string, headers map[string]string, retry int) ([]byte, error) {
	req := fasthttp.AcquireRequest()
	req.SetRequestURI(url)

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/83.0.4103.61 Safari/537.36")
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)

	for i := 0; i < retry; i++ {
		if err := fasthttp.Do(req, resp); err != nil {
			continue
		}

		body := resp.Body()
		code := resp.StatusCode()
		if code != fasthttp.StatusOK || len(body) == 0 {
			return nil, errors.New("http code:" + strconv.Itoa(code))
		}

		return body, nil
	}
	return nil, errors.New("fail with request: " + url)
}
