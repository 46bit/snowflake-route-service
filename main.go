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
		doc.Find("body").AppendHtml("<script>alert('appended by the snowflake-route-service')</script>")
		htmlString, err := doc.Html()
		if err != nil {
			log.Printf("unable to get HTML of modified doc: %s\n", err)
			return err
		}
		html = []byte(htmlString)
	}

	response.Body = ioutil.NopCloser(bytes.NewReader(html))
	return nil
}
