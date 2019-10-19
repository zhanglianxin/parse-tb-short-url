package main

import (
	"flag"
	"os"
	"github.com/valyala/fasthttp"
	"fmt"
	"regexp"
	"sync"
)

const mUA = "Mozilla/5.0 (iPhone; CPU iPhone OS 10_2 like Mac OS X) AppleWebKit/602.3.12 (KHTML, like Gecko) Mobile/14C92 ChannelId(3) Nebula PSDType(1) AlipayDefined(nt:WIFI,ws:375|647|2.0) AliApp(AP/10.0.1.123008) AlipayClient/10.0.1.123008 Alipay Language/zh-Hans"
const rawUrlTmpl = "https://item.taobao.com/item.htm?id=%s"

var (
	h      bool
	ua     string
	urlReg = "\\bhttps:\\/\\/(m\\.tb\\.cn|c\\.tb\\.cn)/[\\w\\&\\=\\%\\.\\:\\;\\?]+"
	// i568798650647.htm
	mTaobaoReg = "https\\:\\/\\/a\\.m\\.taobao\\.com\\/i(\\d+)\\.htm\\?.+\\b"
	// spm=a21wq.8999005.603891285329.2
	sClickReg = "https\\:\\/\\/s\\.click\\.taobao\\.com\\/t\\?e\\=.+spm\\=\\w+\\.\\d+\\.(\\d+)\\.\\d.+\\b"
	// spm=a21wq.8999005.602362101964.2
	uLandReg = "https\\:\\/\\/uland\\.taobao\\.com\\/coupon\\/edetail\\?e\\=.+spm\\=\\w+\\.\\d+\\.(\\d+)\\.\\d.+\\b"
)

func init() {
	flag.BoolVar(&h, "h", false, "this help")
	flag.StringVar(&ua, "ua", mUA, "set user-agent")
	flag.Usage = usage
}

func main() {
	flag.Parse()
	args := flag.Args()
	if h || 0 == len(args) {
		flag.Usage()
		os.Exit(1)
	}

	var wg sync.WaitGroup
	wg.Add(len(args))
	for _, url := range args {
		go func(url string) {
			defer wg.Done()
			if matched, _ := regexp.MatchString(urlReg, url); !matched {
				fmt.Fprintf(os.Stderr, "url[%s] is not supported\n", url)
				return
			}
			req := fasthttp.AcquireRequest()
			res := fasthttp.AcquireResponse()
			defer func() {
				fasthttp.ReleaseResponse(res)
				fasthttp.ReleaseRequest(req)
			}()

			req.Header.SetUserAgent(ua)
			req.Header.SetMethod("GET")
			req.SetRequestURI(url)
			if err := fasthttp.Do(req, res); nil != err {
				panic(err)
			}
			content := string(res.Body())
			matches := findIdFromStr(mTaobaoReg, content)
			if nil == matches {
				matches = findIdFromStr(sClickReg, content)
				if nil == matches {
					matches = findIdFromStr(uLandReg, content)
					if nil == matches {
						fmt.Fprintf(os.Stderr, "url[%s] is not supported\n", url)
					} else {
						id := matches[1]
						realUrl := getRealUrlById(id)
						fmt.Println("uland", id, realUrl)
					}
				} else {
					id := matches[1]
					realUrl := getRealUrlById(id)
					fmt.Println("s.click", id, realUrl)
				}
			} else {
				id := matches[1]
				realUrl := getRealUrlById(id)
				fmt.Println("a.m", id, realUrl)
			}
		}(url)
	}
	wg.Wait()
}

func usage() {
	fmt.Fprintf(os.Stdout, "Usage: %s [options...] <url>\n\nOptions:\n", os.Args[0])
	flag.PrintDefaults()
}

func findIdFromStr(reg string, s string) ([]string) {
	re, _ := regexp.Compile(reg)
	return re.FindStringSubmatch(s)
}

func getRealUrlById(id string) (string) {
	req := fasthttp.AcquireRequest()
	res := fasthttp.AcquireResponse()
	defer func() {
		fasthttp.ReleaseResponse(res)
		fasthttp.ReleaseRequest(req)
	}()

	req.Header.SetUserAgent(ua)
	req.Header.SetMethod("HEAD")
	url := fmt.Sprintf(rawUrlTmpl, id)
	req.SetRequestURI(url)
	if err := fasthttp.Do(req, res); nil != err {
		panic(err)
	}
	statusCode := res.StatusCode()
	if fasthttp.StatusOK == statusCode {
		return url
	} else if fasthttp.StatusMovedPermanently == statusCode || fasthttp.StatusFound == statusCode {
		location := res.Header.Peek("location")
		if nil != location {
			return string(location)
		} else {
			return ""
		}
	} else {
		return ""
	}
}
