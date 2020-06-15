package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/elazarl/goproxy"
)

var privateIPBlocks []*net.IPNet

func init() {
	for _, cidr := range []string{
		"127.0.0.0/8",    // IPv4 loopback
		"10.0.0.0/8",     // RFC1918
		"172.16.0.0/12",  // RFC1918
		"192.168.0.0/16", // RFC1918
		"169.254.0.0/16", // RFC3927 link-local
		"::1/128",        // IPv6 loopback
		"fe80::/10",      // IPv6 link-local
		"fc00::/7",       // IPv6 unique local addr
	} {
		_, block, err := net.ParseCIDR(cidr)
		if err != nil {
			panic(fmt.Errorf("parse error on %q: %v", cidr, err))
		}
		privateIPBlocks = append(privateIPBlocks, block)
	}
}

func isPrivateIP(ip net.IP) bool {
	if ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
		return true
	}
	for _, block := range privateIPBlocks {
		if block.Contains(ip) {
			return true
		}
	}
	return false
}

func isPrivateHost(host string) (bool, error) {
	ipList, err := net.LookupIP(host)
	if err != nil {
		return false, err
	}
	for _, ip := range ipList {
		if isPrivateIP(ip) {
			return true, nil
		}
	}
	return false, nil
}

func main() {
	verbose := flag.Bool("v", false, "should every proxy request be logged to stdout")
	addr := flag.String("addr", ":8080", "proxy listen address")
	flag.Parse()

	proxy := goproxy.NewProxyHttpServer()
	proxy.Verbose = *verbose

	proxy.OnRequest().DoFunc(func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
		isPrivate, err := isPrivateHost(req.URL.Hostname())
		if err != nil {
			log.Println(err)
			return nil, goproxy.NewResponse(req, goproxy.ContentTypeText, http.StatusBadGateway, "Host not found")
		}
		if isPrivate {
			log.Println("Reject host", req.URL.Hostname(), "that resolves to private IP")
			return nil, goproxy.NewResponse(req, goproxy.ContentTypeText, http.StatusForbidden, "Private IP is not allowed")
		}

		return req, nil
	})

	proxy.OnRequest().HandleConnectFunc(func(hostport string, ctx *goproxy.ProxyCtx) (*goproxy.ConnectAction, string) {
		host, _, err := net.SplitHostPort(hostport)
		if err != nil {
			log.Println(err)
			return goproxy.RejectConnect, hostport
		}

		isPrivate, err := isPrivateHost(host)
		if err != nil {
			log.Println(err)
			return goproxy.RejectConnect, hostport
		}
		if isPrivate {
			log.Println("Reject host", hostport, "that resolves to private IP")
			return goproxy.RejectConnect, hostport
		}

		return goproxy.OkConnect, hostport
	})

	log.Fatal(http.ListenAndServe(*addr, proxy))
}
