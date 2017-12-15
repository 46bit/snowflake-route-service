package main

import (
	"bytes"
	"crypto/tls"
	"github.com/PuerkitoBio/goquery"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
)

func main() {
	skipSslValidation, err := strconv.ParseBool(os.Getenv("SKIP_SSL_VALIDATION"))
	if err != nil {
		skipSslValidation = true
	}

	http.Handle("/", snowflakeProxy(skipSslValidation))
	log.Fatalln(http.ListenAndServe(":"+os.Getenv("PORT"), nil))
}

func snowflakeProxy(skipSslValidation bool) http.Handler {
	configuredTransport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: skipSslValidation},
	}
	return &httputil.ReverseProxy{
		Director:       forwardingDirector,
		Transport:      configuredTransport,
		ModifyResponse: applySnowflakes,
	}
}

func forwardingDirector(req *http.Request) {
	forwardedURLString := req.Header.Get("X-Cf-Forwarded-Url")
	forwardedURL, err := url.Parse(forwardedURLString)
	if err != nil {
		log.Printf("unable to parse forwarded url '%s': %s\n", forwardedURLString, err)
		return
	}
	req.URL = forwardedURL
	req.Host = forwardedURL.Host
}

func applySnowflakes(response *http.Response) error {
	html, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Printf("unable to read body of response: %s\n", err)
		return err
	}

	bodyReader := bytes.NewBuffer(html)
	doc, err := goquery.NewDocumentFromReader(bodyReader)
	if err == nil {
		doc.Find("body").AppendHtml(snowHTML)
		htmlString, err := doc.Html()
		if err != nil {
			log.Printf("unable to get HTML of modified doc: %s\n", err)
			return err
		}
		html = []byte(htmlString)
	}

	response.Body = ioutil.NopCloser(bytes.NewReader(html))
	response.ContentLength = int64(len(html))
	if _, ok := response.Header["Content-Length"]; ok {
		response.Header["Content-Length"][0] = string(len(html))
	}

	return nil
}

// https://github.com/pajasevi/CSSnowflakes used under its MIT License
const (
	snowHTML = `
<style>
/* customizable snowflake styling */
.snowflake {
  color: #d7fbff;
  font-size: 2em;
  font-family: Arial;
  text-shadow: 0 0 1px #000;
}
@-webkit-keyframes snowflakes-fall{0%{top:-10%}100%{top:100%}}@-webkit-keyframes snowflakes-shake{0%{-webkit-transform:translateX(0px);transform:translateX(0px)}50%{-webkit-transform:translateX(80px);transform:translateX(80px)}100%{-webkit-transform:translateX(0px);transform:translateX(0px)}}@keyframes snowflakes-fall{0%{top:-10%}100%{top:100%}}@keyframes snowflakes-shake{0%{transform:translateX(0px)}50%{transform:translateX(80px)}100%{transform:translateX(0px)}}.snowflake{position:fixed;top:-10%;z-index:9999;-webkit-user-select:none;-moz-user-select:none;-ms-user-select:none;user-select:none;cursor:default;-webkit-animation-name:snowflakes-fall,snowflakes-shake;-webkit-animation-duration:10s,3s;-webkit-animation-timing-function:linear,ease-in-out;-webkit-animation-iteration-count:infinite,infinite;-webkit-animation-play-state:running,running;animation-name:snowflakes-fall,snowflakes-shake;animation-duration:10s,3s;animation-timing-function:linear,ease-in-out;animation-iteration-count:infinite,infinite;animation-play-state:running,running}.snowflake:nth-of-type(0){left:1%;-webkit-animation-delay:0s,0s;animation-delay:0s,0s}.snowflake:nth-of-type(1){left:10%;-webkit-animation-delay:1s,1s;animation-delay:1s,1s}.snowflake:nth-of-type(2){left:20%;-webkit-animation-delay:6s,.5s;animation-delay:6s,.5s}.snowflake:nth-of-type(3){left:30%;-webkit-animation-delay:4s,2s;animation-delay:4s,2s}.snowflake:nth-of-type(4){left:40%;-webkit-animation-delay:2s,2s;animation-delay:2s,2s}.snowflake:nth-of-type(5){left:50%;-webkit-animation-delay:8s,3s;animation-delay:8s,3s}.snowflake:nth-of-type(6){left:60%;-webkit-animation-delay:6s,2s;animation-delay:6s,2s}.snowflake:nth-of-type(7){left:70%;-webkit-animation-delay:2.5s,1s;animation-delay:2.5s,1s}.snowflake:nth-of-type(8){left:80%;-webkit-animation-delay:1s,0s;animation-delay:1s,0s}.snowflake:nth-of-type(9){left:90%;-webkit-animation-delay:3s,1.5s;animation-delay:3s,1.5s}
</style>
<div class="snowflakes" aria-hidden="true">
  <div class="snowflake">❄</div>
  <div class="snowflake">❅</div>
  <div class="snowflake">❆</div>
  <div class="snowflake">❄</div>
  <div class="snowflake">❅</div>
  <div class="snowflake">❆</div>
  <div class="snowflake">❄</div>
  <div class="snowflake">❅</div>
  <div class="snowflake">❆</div>
  <div class="snowflake">❄</div>
</div>
`
)
