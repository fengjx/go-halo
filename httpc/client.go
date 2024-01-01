package httpc

import (
	"bytes"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/fengjx/go-halo/json"
)

const (
	// MethodGet HTTP method
	MethodGet = http.MethodGet

	// MethodPost HTTP method
	MethodPost = http.MethodPost

	// MethodPut HTTP method
	MethodPut = http.MethodPut

	// MethodDelete HTTP method
	MethodDelete = http.MethodDelete

	// MethodPatch HTTP method
	MethodPatch = http.MethodPatch

	// MethodHead HTTP method
	MethodHead = http.MethodHead

	// MethodOptions HTTP method
	MethodOptions = http.MethodOptions

	plainTextType   = "text/plain; charset=utf-8"
	jsonContentType = "application/json"
	formContentType = "application/x-www-form-urlencoded"
)

var (
	hdrUserAgentKey       = http.CanonicalHeaderKey("User-Agent")
	hdrAcceptKey          = http.CanonicalHeaderKey("Accept")
	hdrContentTypeKey     = http.CanonicalHeaderKey("Content-Type")
	hdrContentLengthKey   = http.CanonicalHeaderKey("Content-Length")
	hdrContentEncodingKey = http.CanonicalHeaderKey("Content-Encoding")
	hdrLocationKey        = http.CanonicalHeaderKey("Location")
	hdrAuthorizationKey   = http.CanonicalHeaderKey("Authorization")
	hdrWwwAuthenticateKey = http.CanonicalHeaderKey("WWW-Authenticate")
)

type Client struct {
	*http.Client
}

type Config struct {
	BaseURL        string
	Timeout        time.Duration
	DefaultHeaders map[string]string
	Transport      http.RoundTripper
	CheckRedirect  func(req *http.Request, via []*http.Request) error
	Jar            http.CookieJar
}

type TransportWrap struct {
	Headers map[string]string
	T       http.RoundTripper
}

func (t *TransportWrap) RoundTrip(req *http.Request) (*http.Response, error) {
	for name, value := range t.Headers {
		req.Header.Add(name, value)
	}
	return t.T.RoundTrip(req)
}

func New(config *Config) *Client {
	defaultTranspor := http.DefaultTransport
	if config.Transport != nil {
		defaultTranspor = config.Transport
	}
	transport := &TransportWrap{
		Headers: config.DefaultHeaders,
		T:       defaultTranspor,
	}
	cli := &http.Client{
		Timeout:       config.Timeout,
		CheckRedirect: config.CheckRedirect,
		Jar:           config.Jar,
		Transport:     transport,
	}
	return &Client{
		Client: cli,
	}
}

func (cli *Client) Request(req *http.Request) (*Response, error) {
	start := time.Now()
	resp, err := cli.Do(req)
	response := &Response{
		sendAt:      start,
		RawResponse: resp,
	}
	if err != nil {
		response.setReceivedAt()
		return response, err
	}
	defer closeq(resp.Body)
	body := resp.Body
	if response.body, err = io.ReadAll(body); err != nil {
		response.setReceivedAt()
		return response, err
	}
	response.setReceivedAt()
	response.size = int64(len(response.body))
	return response, nil
}

func (cli *Client) Get(requestURL string, params map[string]string) (*Response, error) {
	req, err := http.NewRequest(MethodGet, requestURL, nil)
	if err != nil {
		return nil, err
	}
	for name, value := range params {
		req.URL.Query().Add(name, value)
	}
	return cli.Request(req)
}

func (cli *Client) Post(requestURL string, data interface{}) (resp *Response, err error) {
	var req *http.Request
	if data == nil {
		req, err = http.NewRequest(MethodPost, requestURL, nil)
	} else {
		var bys []byte
		bys, err = json.ToBytes(data)
		if err != nil {
			return nil, err
		}
		bodyBuf := bytes.NewBuffer(bys)
		req, err = http.NewRequest(MethodPost, requestURL, bodyBuf)
	}
	if err != nil {
		return nil, err
	}
	req.Header.Set(hdrContentTypeKey, jsonContentType)
	return cli.Request(req)
}

func (cli *Client) PostForm(requestURL string, formData map[string]string) (resp *Response, err error) {
	var req *http.Request
	if len(formData) == 0 {
		req, err = http.NewRequest(MethodPost, requestURL, nil)
	} else {
		formDataValues := url.Values{}
		for k, v := range formData {
			formDataValues.Add(k, v)
		}
		bodyBuf := bytes.NewBuffer([]byte(formDataValues.Encode()))
		req, err = http.NewRequest(MethodPost, requestURL, bodyBuf)
	}
	if err != nil {
		return nil, err
	}
	req.Header.Set(hdrContentTypeKey, formContentType)
	return cli.Request(req)
}

func closeq(v interface{}) {
	if c, ok := v.(io.Closer); ok {
		silently(c.Close())
	}
}

func silently(_ ...interface{}) {}
