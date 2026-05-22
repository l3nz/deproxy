package main

import (
	"crypto/x509"
	"encoding/pem"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/elazarl/goproxy"
)

func main() {
	addr := flag.String("addr", ":8080", "proxy listen address")
	verbose := flag.Bool("v", false, "verbose logging")
	dumpCert := flag.String("dump-cert", "", "write CA cert to this file and exit (import into Firefox to trust HTTPS interception)")
	flag.Parse()

	if *dumpCert != "" {
		cert, err := x509.ParseCertificate(goproxy.GoproxyCa.Certificate[0])
		if err != nil {
			log.Fatal(err)
		}
		f, err := os.Create(*dumpCert)
		if err != nil {
			log.Fatal(err)
		}
		pem.Encode(f, &pem.Block{Type: "CERTIFICATE", Bytes: cert.Raw})
		f.Close()
		fmt.Printf("CA cert written to %s\n", *dumpCert)
		return
	}

	proxy := goproxy.NewProxyHttpServer()
	proxy.Verbose = *verbose

	// Log every request.
	proxy.OnRequest().DoFunc(func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
		log.Printf("-> %s %s", req.Method, req.URL)
		ctx.UserData = time.Now()
		return req, nil
	})

	// Log every response.
	proxy.OnResponse().DoFunc(func(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
		if resp != nil {
			elapsed := time.Since(ctx.UserData.(time.Time))
			ct := resp.Header.Get("Content-Type")
			size := max(resp.ContentLength, 0)
			log.Printf("<- %d %s ct=%s size=%d ms=%d", resp.StatusCode, ctx.Req.URL, ct, size, elapsed.Milliseconds())
		}
		return resp
	})

	// Example: replace a specific request with a local response.
	// Modify the condition below to match whichever URL you want to intercept.
	proxy.OnRequest(matchHost("example.com")).DoFunc(func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
		body := "intercepted by deproxy"
		resp := goproxy.NewResponse(req, "text/plain", http.StatusOK, body)
		log.Printf("!! intercepted %s", req.URL)
		return req, resp
	})

	// Handle CONNECT (HTTPS) tunneling — required for HTTPS interception.
	proxy.OnRequest().HandleConnect(goproxy.AlwaysMitm)

	fmt.Printf("deproxy listening on %s\n", *addr)
	log.Fatal(http.ListenAndServe(*addr, proxy))
}

// matchHost returns a goproxy condition that matches requests whose host
// contains the given substring.
func matchHost(substring string) goproxy.ReqConditionFunc {
	return func(req *http.Request, ctx *goproxy.ProxyCtx) bool {
		return strings.Contains(req.URL.Host, substring)
	}
}
