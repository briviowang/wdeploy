package runtime

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

// HTTPService http类
type HTTPService struct {
	Verbose  bool
	ProxyURL string
	Headers  map[string][]string
}

// HTTPServiceDownloadResult http下载结果
type HTTPServiceDownloadResult struct {
	ErrorText string
	FilePath  string
	FileSize  int64
	URL       string
	Duration  float64
}

// NewHTTPService 实例化
func NewHTTPService() HTTPService {
	return HTTPService{
		Verbose: true,
	}
}

func (c *HTTPService) setVerbose(v bool) {
	c.Verbose = v
}

func (c *HTTPService) getHTTPClient() *http.Client {
	t := &http.Transport{
		Dial: (&net.Dialer{
			Timeout:   60 * time.Second,
			KeepAlive: 30 * time.Second,
		}).Dial,
		TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
		TLSHandshakeTimeout: 60 * time.Second,
	}

	if IsEmpty(c.ProxyURL) {
		c.ProxyURL = os.Getenv("http_proxy")
	}

	if IsURL(c.ProxyURL) {
		proxyURL, _ := url.Parse(c.ProxyURL)
		t.Proxy = http.ProxyURL(proxyURL)
	}

	client := &http.Client{
		Transport: t,
	}
	return client
}

func (c *HTTPService) getHTTPRequest(method, url string, body io.Reader) *http.Request {
	request, _ := http.NewRequest(method, url, body)

	request.Header = c.Headers
	request.Close = false
	return request
}

func (c *HTTPService) executeHTTPRequest(req *http.Request, tryLimit int) *http.Response {
	client := c.getHTTPClient()

	for i := 0; i < tryLimit; i++ {
		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("%v\n", err)
		}
		if err == nil && resp != nil && resp.Body != nil {
			return resp
		}
	}
	return nil
}

// Get get请求
func (c *HTTPService) Get(URL string, params map[string]string) string {
	req := c.getHTTPRequest("GET", URL, nil)

	q := req.URL.Query()
	for key, val := range params {
		q.Add(key, val)
	}

	req.URL.RawQuery = q.Encode()

	fmt.Println(req.URL.String())
	client := http.Client{}
	resp, _ := client.Do(req)
	body, _ := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	return string(body)
}

// Post post请求
func (c *HTTPService) Post(URL string, params map[string]string) string {
	temp := []string{}
	for key, val := range params {
		temp = append(temp, key+"="+val)
	}

	req := c.getHTTPRequest("POST", URL, strings.NewReader(Implode("&", temp)))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	// fmt.Println(JSONEncode(req.Header))

	client := c.getHTTPClient()
	resp, _ := client.Do(req)
	body, _ := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	return string(body)
}

func (c *HTTPService) printDownloadPercent(done chan int64, path string, total int64) {
	var stop = false
	for {
		select {
		case <-done:
			stop = true
		default:
			file, err := os.Open(path)
			if err != nil {
				log.Fatal(err)
			}
			defer file.Close()

			fi, err := file.Stat()
			if err != nil {
				log.Fatal(err)
			}

			size := fi.Size()

			if size == 0 {
				size = 1
			}

			if c.Verbose {
				if total <= 0 {
					fmt.Printf("Downloading: %-10s\r", FormatSize(size))
				} else {
					var percent = float64(size) / float64(total) * 100
					fmt.Printf("Downloading: %.2f%s\r", percent, "%")
				}
			}
		}

		if stop {
			break
		}

		time.Sleep(time.Second)
	}
}

// Download 下载文件
//HTTPService := HTTPService{}
//HTTPService.Download("https://wordpress.org/wordpress-4.4.2.zip", "./")
func (c *HTTPService) Download(url string, dest string) HTTPServiceDownloadResult {
	result := HTTPServiceDownloadResult{}
	result.URL = url
	var p bytes.Buffer
	p.WriteString(dest)

	println(url)

	if !IsURL(url) {
		result.ErrorText = "url不合法"
		return result
	}
	if c.Verbose {
		fmt.Printf("Url:%s\n", url)
	}

	res := Explode(".", BaseName(dest))
	if len(res) < 2 {
		p.WriteString("/")
		p.WriteString(path.Base(url))
	}
	filePath := p.String()
	result.FilePath = filePath

	fileSize := GetFileSize(filePath)
	if fileSize > 0 {
		result.FileSize = fileSize
		result.ErrorText = "文件已经存在"
		return result
	}
	start := time.Now()

	out, err := os.Create(filePath)

	if err != nil {
		println(filePath)
		panic(err)
	}

	headRequest := c.getHTTPRequest("HEAD", url, nil)
	headResp := c.executeHTTPRequest(headRequest, 3)

	if headResp == nil {
		result.ErrorText = "响应头为空"
		return result
	}

	defer headResp.Body.Close()

	size := 0
	contentLength, err := strconv.Atoi(headResp.Header.Get("Content-Length"))
	if err != nil {
		size = contentLength
	}

	done := make(chan int64)
	go c.printDownloadPercent(done, filePath, int64(size))

	downloadRequest := c.getHTTPRequest("GET", url, nil)

	resp := c.executeHTTPRequest(downloadRequest, 3)
	if resp == nil {
		result.ErrorText = "返回内容为空"
		return result
	}

	n, err := io.Copy(out, resp.Body)

	if err != nil {
		fmt.Printf(err.Error())
	}
	out.Close()
	resp.Body.Close()

	done <- n
	if c.Verbose {
		fmt.Printf("Downloaded:%-10s\r", FormatSize(GetFileSize(filePath)))
		fmt.Println("")
		elapsed := time.Since(start)
		fmt.Printf("Download completed in %.2fs\n", elapsed.Seconds())
	}
	result.FileSize = GetFileSize(filePath)

	seconds, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", time.Since(start).Seconds()), 64)
	result.Duration = seconds
	result.ErrorText = fmt.Sprintf("大小:%s,耗时:%.2fs", FormatSize(result.FileSize), result.Duration)
	return result
}
