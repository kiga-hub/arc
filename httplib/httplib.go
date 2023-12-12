package httplib

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/tls"
	"encoding/json"
	"encoding/xml"
	"io"
	"log"
	"mime/multipart"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/http/httputil"
	"net/url"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"gopkg.in/yaml.v2"
)

var defaultSetting = HTTPSettings{
	UserAgent:        "DeviceServer",
	ConnectTimeout:   60 * time.Second,
	ReadWriteTimeout: 60 * time.Second,
	Gzip:             true,
	DumpBody:         true,
}

var defaultCookieJar http.CookieJar
var settingMutex sync.Mutex

// createDefaultCookie creates a global cookiejar to store cookies.
func createDefaultCookie() {
	settingMutex.Lock()
	defer settingMutex.Unlock()
	defaultCookieJar, _ = cookiejar.New(nil)
}

// SetDefaultSetting Overwrite default settings
//
//goland:noinspection GoUnusedExportedFunction
func SetDefaultSetting(setting HTTPSettings) {
	settingMutex.Lock()
	defer settingMutex.Unlock()
	defaultSetting = setting
}

// NewRequest return *BeegoHttpRequest with specific method
func NewRequest(rawURL, method string) *HTTPRequest {
	var resp http.Response
	u, err := url.Parse(rawURL)
	if err != nil {
		log.Println("HTTP lib:", err)
	}
	req := http.Request{
		URL:        u,
		Method:     method,
		Header:     make(http.Header),
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
	}
	return &HTTPRequest{
		url:     rawURL,
		req:     &req,
		params:  map[string][]string{},
		files:   map[string]string{},
		setting: defaultSetting,
		resp:    &resp,
	}
}

// Get returns *HttpRequest with GET method.
//
//goland:noinspection GoUnusedExportedFunction
func Get(url string) *HTTPRequest {
	return NewRequest(url, "GET")
}

// Post returns *HttpRequest with POST method.
//
//goland:noinspection GoUnusedExportedFunction
func Post(url string) *HTTPRequest {
	return NewRequest(url, "POST")
}

// Put returns *HttpRequest with PUT method.
//
//goland:noinspection GoUnusedExportedFunction
func Put(url string) *HTTPRequest {
	return NewRequest(url, "PUT")
}

// Delete returns *HttpRequest DELETE method.
//
//goland:noinspection GoUnusedExportedFunction
func Delete(url string) *HTTPRequest {
	return NewRequest(url, "DELETE")
}

// Head returns *HttpRequest with HEAD method.
//
//goland:noinspection GoUnusedExportedFunction
func Head(url string) *HTTPRequest {
	return NewRequest(url, "HEAD")
}

// HTTPSettings is the http.Client setting
type HTTPSettings struct {
	ShowDebug        bool
	UserAgent        string
	ConnectTimeout   time.Duration
	ReadWriteTimeout time.Duration
	TLSClientConfig  *tls.Config
	Proxy            func(*http.Request) (*url.URL, error)
	Transport        http.RoundTripper
	CheckRedirect    func(req *http.Request, via []*http.Request) error
	EnableCookie     bool
	Gzip             bool
	DumpBody         bool
	Retries          int // if set to -1 means will retry forever
	RetryDelay       time.Duration
}

// HTTPRequest provides more useful methods for requesting one url than http.Request.
type HTTPRequest struct {
	url     string
	req     *http.Request
	params  map[string][]string
	files   map[string]string
	setting HTTPSettings
	resp    *http.Response
	body    []byte
	dump    []byte
}

// GetRequest return the request object
func (b *HTTPRequest) GetRequest() *http.Request {
	return b.req
}

// Setting Change request settings
func (b *HTTPRequest) Setting(setting HTTPSettings) *HTTPRequest {
	b.setting = setting
	return b
}

// SetBasicAuth sets the request's Authorization header to use HTTP Basic Authentication with the provided username and password.
func (b *HTTPRequest) SetBasicAuth(username, password string) *HTTPRequest {
	b.req.SetBasicAuth(username, password)
	return b
}

// SetEnableCookie sets enable/disable cookiejar
func (b *HTTPRequest) SetEnableCookie(enable bool) *HTTPRequest {
	b.setting.EnableCookie = enable
	return b
}

// SetUserAgent sets User-Agent header field
func (b *HTTPRequest) SetUserAgent(useragent string) *HTTPRequest {
	b.setting.UserAgent = useragent
	return b
}

// Debug sets show debug or not when executing request.
func (b *HTTPRequest) Debug(isDebug bool) *HTTPRequest {
	b.setting.ShowDebug = isDebug
	return b
}

// Retries sets Retries times.
// default is 0 means no retried.
// -1 means retried forever.
// other means retried times.
func (b *HTTPRequest) Retries(times int) *HTTPRequest {
	b.setting.Retries = times
	return b
}

// RetryDelay -
func (b *HTTPRequest) RetryDelay(delay time.Duration) *HTTPRequest {
	b.setting.RetryDelay = delay
	return b
}

// DumpBody setting whether you need to Dump the Body.
func (b *HTTPRequest) DumpBody(isDump bool) *HTTPRequest {
	b.setting.DumpBody = isDump
	return b
}

// DumpRequest return the DumpRequest
func (b *HTTPRequest) DumpRequest() []byte {
	return b.dump
}

// SetTimeout sets connect time out and read-write time out for BeegoRequest.
func (b *HTTPRequest) SetTimeout(connectTimeout, readWriteTimeout time.Duration) *HTTPRequest {
	b.setting.ConnectTimeout = connectTimeout
	b.setting.ReadWriteTimeout = readWriteTimeout
	return b
}

// SetTLSClientConfig sets tls connection configurations if visiting https url.
func (b *HTTPRequest) SetTLSClientConfig(config *tls.Config) *HTTPRequest {
	b.setting.TLSClientConfig = config
	return b
}

// Header add header item string in request.
func (b *HTTPRequest) Header(key, value string) *HTTPRequest {
	b.req.Header.Set(key, value)
	return b
}

// SetHost set the request host
func (b *HTTPRequest) SetHost(host string) *HTTPRequest {
	b.req.Host = host
	return b
}

// SetProtocolVersion Set the protocol version for incoming requests.
// Client requests always use HTTP/1.1.
func (b *HTTPRequest) SetProtocolVersion(version string) *HTTPRequest {
	if len(version) == 0 {
		version = "HTTP/1.1"
	}

	major, minor, ok := http.ParseHTTPVersion(version)
	if ok {
		b.req.Proto = version
		b.req.ProtoMajor = major
		b.req.ProtoMinor = minor
	}

	return b
}

// SetCookie add cookie into request.
func (b *HTTPRequest) SetCookie(cookie *http.Cookie) *HTTPRequest {
	b.req.Header.Add("Cookie", cookie.String())
	return b
}

// SetTransport set the setting transport
func (b *HTTPRequest) SetTransport(transport http.RoundTripper) *HTTPRequest {
	b.setting.Transport = transport
	return b
}

// SetProxy set the http proxy
// example:
//
//	func(req *http.Request) (*url.URL, error) {
//		u, _ := url.ParseRequestURI("http://127.0.0.1:8118")
//		return u, nil
//	}
func (b *HTTPRequest) SetProxy(proxy func(*http.Request) (*url.URL, error)) *HTTPRequest {
	b.setting.Proxy = proxy
	return b
}

// SetCheckRedirect specifies the policy for handling redirects.
//
// If CheckRedirect is nil, the Client uses its default policy,
// which is to stop after 10 consecutive requests.
func (b *HTTPRequest) SetCheckRedirect(redirect func(req *http.Request, via []*http.Request) error) *HTTPRequest {
	b.setting.CheckRedirect = redirect
	return b
}

// Param adds query param in to request.
// params build query string as ?key1=value1&key2=value2...
func (b *HTTPRequest) Param(key, value string) *HTTPRequest {
	if param, ok := b.params[key]; ok {
		b.params[key] = append(param, value)
	} else {
		b.params[key] = []string{value}
	}
	return b
}

// PostFile add a post file to the request
func (b *HTTPRequest) PostFile(formName, filename string) *HTTPRequest {
	b.files[formName] = filename
	return b
}

// Body adds request raw body.
// it supports string and []byte.
func (b *HTTPRequest) Body(data interface{}) *HTTPRequest {
	switch t := data.(type) {
	case string:
		bf := bytes.NewBufferString(t)
		b.req.Body = io.NopCloser(bf)
		b.req.ContentLength = int64(len(t))
	case []byte:
		bf := bytes.NewBuffer(t)
		b.req.Body = io.NopCloser(bf)
		b.req.ContentLength = int64(len(t))
	}
	return b
}

// XMLBody adds request raw body encoding by XML.
func (b *HTTPRequest) XMLBody(obj interface{}) (*HTTPRequest, error) {
	if b.req.Body == nil && obj != nil {
		data, err := xml.Marshal(obj)
		if err != nil {
			return b, err
		}
		b.req.Body = io.NopCloser(bytes.NewReader(data))
		b.req.ContentLength = int64(len(data))
		b.req.Header.Set("Content-Type", "application/xml")
	}
	return b, nil
}

// YAMLBody adds request raw body encoding by YAML.
func (b *HTTPRequest) YAMLBody(obj interface{}) (*HTTPRequest, error) {
	if b.req.Body == nil && obj != nil {
		data, err := yaml.Marshal(obj)
		if err != nil {
			return b, err
		}
		b.req.Body = io.NopCloser(bytes.NewReader(data))
		b.req.ContentLength = int64(len(data))
		b.req.Header.Set("Content-Type", "application/x+yaml")
	}
	return b, nil
}

// JSONBody adds request raw body encoding by JSON.
func (b *HTTPRequest) JSONBody(obj interface{}) (*HTTPRequest, error) {
	if b.req.Body == nil && obj != nil {
		data, err := json.Marshal(obj)
		if err != nil {
			return b, err
		}
		b.req.Body = io.NopCloser(bytes.NewReader(data))
		b.req.ContentLength = int64(len(data))
		b.req.Header.Set("Content-Type", "application/json")
	}
	return b, nil
}

func (b *HTTPRequest) buildURL(paramBody string) {
	// build GET url with query string
	if b.req.Method == "GET" && len(paramBody) > 0 {
		if strings.Contains(b.url, "?") {
			b.url += "&" + paramBody
		} else {
			b.url = b.url + "?" + paramBody
		}
		return
	}

	// build POST/PUT/PATCH url and body
	if (b.req.Method == "POST" || b.req.Method == "PUT" || b.req.Method == "PATCH" || b.req.Method == "DELETE") && b.req.Body == nil {
		// with files
		if len(b.files) > 0 {
			pr, pw := io.Pipe()
			bodyWriter := multipart.NewWriter(pw)
			go func() {
				for formName, filename := range b.files {
					fileWriter, err := bodyWriter.CreateFormFile(formName, filename)
					if err != nil {
						log.Println("HTTP lib:", err)
					}
					fh, err := os.Open(filename)
					if err != nil {
						log.Println("HTTP lib:", err)
					}
					// io copy
					_, err = io.Copy(fileWriter, fh)
					if err != nil {
						log.Println("HTTP lib:", err)
					}
					err = fh.Close()
					if err != nil {
						log.Println("HTTP lib:", err)
					}
				}
				for k, v := range b.params {
					for _, vv := range v {
						if err := bodyWriter.WriteField(k, vv); err != nil {
							log.Println(err)
						}
					}
				}
				err := bodyWriter.Close()
				if err != nil {
					log.Println("bodyWriter:", err)
				}
				err = pw.Close()
				if err != nil {
					log.Println("pw close:", err)
				}
			}()
			b.Header("Content-Type", bodyWriter.FormDataContentType())
			b.req.Body = io.NopCloser(pr)
			b.Header("Transfer-Encoding", "chunked")
			return
		}

		// with params
		if len(paramBody) > 0 {
			b.Header("Content-Type", "application/x-www-form-urlencoded")
			b.Body(paramBody)
		}
	}
}

func (b *HTTPRequest) getResponse() (*http.Response, error) {
	if b.resp.StatusCode != 0 {
		return b.resp, nil
	}
	resp, err := b.DoRequest()
	if err != nil {
		return nil, err
	}
	b.resp = resp
	return resp, nil
}

// DoRequest will do the client.Do
func (b *HTTPRequest) DoRequest() (resp *http.Response, err error) {
	var paramBody string
	if len(b.params) > 0 {
		var buf bytes.Buffer
		for k, v := range b.params {
			for _, vv := range v {
				buf.WriteString(url.QueryEscape(k))
				buf.WriteByte('=')
				buf.WriteString(url.QueryEscape(vv))
				buf.WriteByte('&')
			}
		}
		paramBody = buf.String()
		paramBody = paramBody[0 : len(paramBody)-1]
	}

	b.buildURL(paramBody)
	urlParsed, err := url.Parse(b.url)
	if err != nil {
		return nil, err
	}

	b.req.URL = urlParsed

	trans := b.setting.Transport

	if trans == nil {
		// create default transport
		trans = &http.Transport{
			TLSClientConfig:     b.setting.TLSClientConfig,
			Proxy:               b.setting.Proxy,
			DialContext:         TimeoutDialer(context.Background(), b.setting.ConnectTimeout, b.setting.ReadWriteTimeout),
			MaxIdleConnsPerHost: 100,
		}
	} else {
		// if b.transport is *http.Transport then set the settings.
		if t, ok := trans.(*http.Transport); ok {
			if t.TLSClientConfig == nil {
				t.TLSClientConfig = b.setting.TLSClientConfig
			}
			if t.Proxy == nil {
				t.Proxy = b.setting.Proxy
			}
			if t.DialContext == nil {
				t.DialContext = TimeoutDialer(context.Background(), b.setting.ConnectTimeout, b.setting.ReadWriteTimeout)
			}
		}
	}

	var jar http.CookieJar
	if b.setting.EnableCookie {
		if defaultCookieJar == nil {
			createDefaultCookie()
		}
		jar = defaultCookieJar
	}

	client := &http.Client{
		Transport: trans,
		Jar:       jar,
	}

	if b.setting.UserAgent != "" && b.req.Header.Get("User-Agent") == "" {
		b.req.Header.Set("User-Agent", b.setting.UserAgent)
	}

	if b.setting.CheckRedirect != nil {
		client.CheckRedirect = b.setting.CheckRedirect
	}

	if b.setting.ShowDebug {
		dump, err := httputil.DumpRequest(b.req, b.setting.DumpBody)
		if err != nil {
			log.Println(err.Error())
		}
		b.dump = dump
	}
	// retries default value is 0, it will run once.
	// retries equal to -1, it will run forever until success
	// retries was set, it will retry fixed times.
	// Sleeps for a 400ms in between calls to reduce spam
	for i := 0; b.setting.Retries == -1 || i <= b.setting.Retries; i++ {
		resp, err = client.Do(b.req)
		if err == nil {
			break
		}
		time.Sleep(b.setting.RetryDelay)
	}
	return resp, err
}

// String returns the body string in response.
// it calls Response inner.
func (b *HTTPRequest) String() (string, error) {
	data, err := b.Bytes()
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// Bytes returns the body []byte in response.
// it calls Response inner.
func (b *HTTPRequest) Bytes() ([]byte, error) {
	if b.body != nil {
		return b.body, nil
	}
	resp, err := b.getResponse()
	if err != nil {
		return nil, err
	}
	if resp.Body == nil {
		return nil, nil
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Println("resp body close:", err)
		}
	}()
	if b.setting.Gzip && resp.Header.Get("Content-Encoding") == "gzip" {
		reader, err := gzip.NewReader(resp.Body)
		if err != nil {
			return nil, err
		}
		b.body, err = io.ReadAll(reader)
		return b.body, err
	}
	b.body, err = io.ReadAll(resp.Body)
	return b.body, err
}

// ToFile saves the body data in response to one file.
// it calls Response inner.
func (b *HTTPRequest) ToFile(filename string) error {
	resp, err := b.getResponse()
	if err != nil {
		return err
	}
	if resp.Body == nil {
		return nil
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Println("resp body close:", err)
		}
	}()
	err = pathExistAndMkdir(filename)
	if err != nil {
		return err
	}
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer func() {
		if err := f.Close(); err != nil {
			log.Println("file close:", err)
		}
	}()
	_, err = io.Copy(f, resp.Body)
	return err
}

// Check that the file directory exists, there is no automatically created
func pathExistAndMkdir(filename string) (err error) {
	filename = path.Dir(filename)
	_, err = os.Stat(filename)
	if err == nil {
		return nil
	}
	if os.IsNotExist(err) {
		err = os.MkdirAll(filename, os.ModePerm)
		if err == nil {
			return nil
		}
	}
	return err
}

// ToJSON returns the map that marshals from the body bytes as json in response .
// it calls Response inner.
func (b *HTTPRequest) ToJSON(v interface{}) error {
	data, err := b.Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

// ToXML returns the map that marshals from the body bytes as xml in response .
// it calls Response inner.
func (b *HTTPRequest) ToXML(v interface{}) error {
	data, err := b.Bytes()
	if err != nil {
		return err
	}
	return xml.Unmarshal(data, v)
}

// ToYAML returns the map that marshals from the body bytes as yaml in response .
// it calls Response inner.
func (b *HTTPRequest) ToYAML(v interface{}) error {
	data, err := b.Bytes()
	if err != nil {
		return err
	}
	return yaml.Unmarshal(data, v)
}

// Response executes request client gets response manually.
func (b *HTTPRequest) Response() (*http.Response, error) {
	return b.getResponse()
}

// TimeoutDialer returns functions of connection dialer with timeout settings for http.Transport Dial field.
func TimeoutDialer(ctx context.Context, cTimeout time.Duration, rwTimeout time.Duration) func(ctx context.Context, net, addr string) (c net.Conn, err error) {
	_ = ctx
	return func(ctx context.Context, netParam, addr string) (net.Conn, error) {
		conn, err := net.DialTimeout(netParam, addr, cTimeout)
		if err != nil {
			return nil, err
		}
		err = conn.SetDeadline(time.Now().Add(rwTimeout))
		return conn, err
	}
}
