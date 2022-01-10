// Copyright (c) 2021 Meng Huang (mhboy@outlook.com)
// This package is licensed under a MIT license that can be found in the LICENSE file.

package websocket

import (
	"crypto/tls"
	"crypto/x509"
	"net/http"
	"sync"
	"testing"
)

func TestWSS(t *testing.T) {
	network := "tcp"
	addr := ":8080"
	Serve := func(conn *Conn) {
		for {
			msg, err := conn.ReadMessage(nil)
			if err != nil {
				break
			}
			conn.WriteMessage(msg)
		}
		conn.Close()
	}

	httpServer := &http.Server{
		Addr:    addr,
		Handler: Handler(Serve),
	}
	l, _ := tls.Listen(network, addr, testTLSConfig())
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		httpServer.Serve(l)
	}()
	{
		conn, err := Dial(network, addr, "/", testClientTLSConfig())
		if err != nil {
			t.Error(err)
		}
		msg := "Hello World"
		if err := conn.WriteMessage([]byte(msg)); err != nil {
			t.Error(err)
		}
		data, err := conn.ReadMessage(nil)
		if err != nil {
			t.Error(err)
		} else if string(data) != msg {
			t.Error(string(data))
		}
		conn.Close()
	}
	{
		_, err := Dial(network, addr, "/", nil)
		if err == nil {
			t.Error()
		}
	}
	httpServer.Close()
	wg.Wait()
}

func testTLSConfig() *tls.Config {
	tlsCert, err := tls.X509KeyPair(testCertPEM, testKeyPEM)
	if err != nil {
		panic(err)
	}
	return &tls.Config{Certificates: []tls.Certificate{tlsCert}}
}

func testSkipVerifyTLSConfig() *tls.Config {
	return &tls.Config{InsecureSkipVerify: true}
}

func testClientTLSConfig() *tls.Config {
	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(testCertPEM) {
		panic("failed to append certificates")
	}
	return &tls.Config{RootCAs: certPool, ServerName: "websocket.hslam.com"}
}

func test1ClientTLSConfig() *tls.Config {
	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(testCertPEM) {
		panic("failed to append certificates")
	}
	return &tls.Config{RootCAs: certPool, ServerName: ""}
}

var testKeyPEM = []byte(`-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEAuGxK16Qu3IcyFb+yCcF4h2Dv7Dd2w2A3pF6iA7WFp08ald3C
+bZqoSzcMPEdHPLJevk4TkWG8Qmas7pltFx/8OPlC5WRkz8p/xVtnsUmGsA3qo5b
1NqXx/WRDypbo/eNZ5RDA0sFvwTD0kyu5KGOODRMfEyrHckl6SOgcfniEwNBs8Iw
1QRMFFh4OFeh850eg0yXzrDXnI5sy8x1f/dXPKvcuctS/ZHNpW31FT3FhyIsDlqD
MGLEI/B8UrAbTF2LUt0OraCjVVXE8+m78FIu2alv3daOIy6jrYv0PVtktJCmYMNf
LjQgLc/n8b/O06xpfeElMKmJF2dWSc6foa30zQIDAQABAoIBAG9dfYhgbaffv//g
JUu81+KwR9FV4NK0TIVmW+FvgQj6PKyZIH8Yh6VSaJjpUNJFTiODUVv6ojT1vsSf
X4EdhmjZxVtMc37+Wobd0rdYh90Ji9PjaVLMuXEXOgR1aKdH+sy8fAcGC69A2lso
0UfgwvfvpOw+g+pVqB3z1JRe+ATQIRfGWEh8T5tjgZKObuKs0kxX2BRVO9zFalsg
G32w8VSVir1c5dJygHCAGuLk3ohncoGfFLoEDsZmHgVg5DWVYANbsnJW0Q3ZVsYp
KnvMKuBMySktl+bH0L8If8yut+I03WHtz0Er9IQyC3GO6dfilHiz0pUQV7+DJjVA
ZXNPJaECgYEA3+ZHNqMNT9VlQjYA+eT9CjcxZoKUVg3Tu+9xMNGZo3OZDkki295d
IwzoD/84Nr+pNcBSxlDvN1AXinimGq0QqWfwP7BbqEyJxFByelF8TF+ms+pePW8v
YWpSn4v3S28zJM92dbwczWRQ9mGLdmEB8VlDxHDPb81GkdHaEV5+ajkCgYEA0t0e
cr9eDQ9g1jhDBWF8NDGLcRPtq9Sl4VCE27U2rsnF3X/4zkXaszSmFlUc4ZcRL1OD
DIc2euz0Ch6C2po7RU+6qGI7UFOk3n5MAjolPsNRB1nOj/OyT1SVnWTROp/Mu2zc
X8w0pEXQlNE02PNW0eHD1tLqcKTse4ZK5VINrzUCgYEAuRGkDYJrL3EJONhgqC5i
Bj6m47/Nku/s8ywxGJQ39YZInilP2gOMYrt5WjewpHh6CkcFZI1jngni239sdSJW
YmDakhpZONzDB3Ujmv2dy5dIuPBho1AzDseOsfhEmaK52JRvq1OpTxC7Z1wrpdb7
fx40yLwiipxX15JpOPAtd+kCgYEA0EvTvyBhJN+TFio/snoJOnnSuCIqfroyHq/u
fia1XNY+2j6HJiSFFM+mXZs4S3RyamDBrMeIvseBjtlzA8SlViObTKi01PW7gHoc
VXrgve4tBejmDvd5pbn1jaRAtvuSP3ca/pr3SWsZz1gWL1W55txxG64AHsQcQy12
oK98ix0CgYB38rvFwiaAvOW6WCMlRl1Yzd9mAK9LI6eygzFV6Ke3lCC4l18XIIRj
KBfLfExkz21WpuY44+LEo3nl7n3xMcHLhIINP8+TaDmhwn3iZBfGUuYkHyMfOTHl
kaQ1jFVjqVCbsyWBSNzterjpeMbxhd/18zzIYnULXGZS++szgSxHsw==
-----END RSA PRIVATE KEY-----
`)

var testCertPEM = []byte(`-----BEGIN CERTIFICATE-----
MIIDSzCCAjOgAwIBAgIURCOmhiGKFKK1oToePEo9e2VuPNMwDQYJKoZIhvcNAQEL
BQAwPzELMAkGA1UEBhMCY24xDjAMBgNVBAsMBW15b3JnMQ8wDQYDVQQKDAZteXRl
c3QxDzANBgNVBAMMBm15bmFtZTAgFw0yMjAxMTAxNTI1MjVaGA8yNTIxMDkxMTE1
MjUyNVowSjELMAkGA1UEBhMCY24xETAPBgNVBAsTCG15c2VydmVyMRMwEQYDVQQK
EwpzZXJ2ZXJjb21wMRMwEQYDVQQDEwpzZXJ2ZXJuYW1lMIIBIjANBgkqhkiG9w0B
AQEFAAOCAQ8AMIIBCgKCAQEAuGxK16Qu3IcyFb+yCcF4h2Dv7Dd2w2A3pF6iA7WF
p08ald3C+bZqoSzcMPEdHPLJevk4TkWG8Qmas7pltFx/8OPlC5WRkz8p/xVtnsUm
GsA3qo5b1NqXx/WRDypbo/eNZ5RDA0sFvwTD0kyu5KGOODRMfEyrHckl6SOgcfni
EwNBs8Iw1QRMFFh4OFeh850eg0yXzrDXnI5sy8x1f/dXPKvcuctS/ZHNpW31FT3F
hyIsDlqDMGLEI/B8UrAbTF2LUt0OraCjVVXE8+m78FIu2alv3daOIy6jrYv0PVtk
tJCmYMNfLjQgLc/n8b/O06xpfeElMKmJF2dWSc6foa30zQIDAQABozIwMDAJBgNV
HRMEAjAAMAsGA1UdDwQEAwIF4DAWBgNVHREEDzANggsqLmhzbGFtLmNvbTANBgkq
hkiG9w0BAQsFAAOCAQEAb3FZTrmqMWzZr0P5mLc5urzIPGlr81xbZ55r6B8kc3aU
jqzr8KISPNAyYxQORIrl+dKe9mCtqRoMfVKkqQ16JahFo/rp/XMYfd/RzgNi3nKh
vrAT/RzOo5+9XhV83PvZJYa2xRqHkh0juT2y6tMFkIEFjIyX+2DEUZ3tkVZscSt+
o6NRuEAWdnyPfAcZMCDwS3hpIuJcVEwqRhqmtxpMwRY9+RMu7nbWgm5E3PfTLqOE
RoJ7VLfEc0IKBHDW6XrY+D5/77AQg7ycDOrV/7i3Ha9JQNrPU/KpOayBg8o4hISL
EJFzAVY7OzZhC50wZjqARgox65xW0Ns4AXClpzPi0Q==
-----END CERTIFICATE-----
`)
