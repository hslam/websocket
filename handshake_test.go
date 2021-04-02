// Copyright (c) 2021 Meng Huang (mhboy@outlook.com)
// This package is licensed under a MIT license that can be found in the LICENSE file.

package websocket

import (
	"crypto/tls"
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
		conn, err := Dial(network, addr, "/", testSkipVerifyTLSConfig())
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

var testKeyPEM = []byte(`-----BEGIN PRIVATE KEY-----
MIIEvAIBADANBgkqhkiG9w0BAQEFAASCBKYwggSiAgEAAoIBAQDa3lfVcQZg3Ra2
DCeTPM9I8cv35Y+R4niXJ7c2U9TvGE3l8zfsLBXtdN4bSlmaimOnOmfx0aVJ8XwL
qcIMspJmzG9UlGdlOfirMTYCybvwhEf9bZc9lmLv27C4++4IljNF9sSv/Lnbdl5V
Nr+lY1xKRR3HPpwwuJj3jh3TznzAnb0QnIRTyGGVThyE6uUQAgx8/taGenJDkzb7
pry4kRvz+GgjAvhi/KOgxho7G6PLfzXeS+iPyaMg5npd3B90XIzaaXr4/yffC5BU
bynHhZLmWKJXSp7brjiZzpFV8np7wKYrtqXW4My2MtMASnvfXrCfTwQ3FU0biBsk
7dQHCcuzAgMBAAECggEAdN6zQhsHT+Pew7j7zOh0uzu6MZYYMssen4Aqmczr7/wn
ZHmaS/dCgjicfTAXZqktC1fptzu+KhzToxqzrroP6OqTLDPOfkQVX7x4XcbBH25T
TqUdVFqgW/oQhMap1VX27Q4W+u5VhDXRq2j/rt2+oz4C56isGGwJ6m6tyLMC9IqJ
Ul9fHrLKKjHltYkCMYzbUP/9QVs9yMlw04BbxCvML21s3ikNuGc8qdQhoHkmxXns
zUR9+P7CkMhSvhojs7MVgaflGozNna89MYAgX+0mCGkWqOXEoFN3n4HdxwW1nBHC
34YndQdOsViO7j9o1SJMOBLMXiQexH+YDJMvjZpsEQKBgQDynyxfINLnHUz1Wo8K
Z1dwmP+dd2av/MbBVsEEyxAugLW8a7Ks6bDlxk8VKB4GAkqzx6Ap+YywZDRJewKn
XUoEG8TPo4dZBy3ttyXTk240zDi/NIJtVRhGxeOGX8zcmwtGHjq694RYeCDrMDWp
yRCJHVUSYUhHwtVwvSZK8JSKCwKBgQDm798FKiqh8UFyuOwKkfdzPYjM6fDg3JuR
E7kmyaeFRz1X0c9zZcWE+ehf9nZnwfU04ZL1WIrjkYWUBxcrBFjChAmFKqFOf97m
0w8jCifuBu+AzaSfW39rzcbpCof9GIHTEczGIjIbj62NQfxhKejP/eA///o81Cf2
hSnUpjn1+QKBgE0jyLLSN9wdl8tmqJYRN17odlU1kmOgBf2QvLvuaE2wxJeM0nlh
r8nOnHRIlgspDWFNtiHCYzXuFiXKw5Q89/yIa7Hs92qZ+sNa+N7lQCPvTpeUdWeX
p6lQ379olDUL4rC/icLKUbzjLOw6HsXF1MkTl2nJnnaafsxih1tKVJ/zAoGAEed8
+fCH96A1u8g8fKFOdv/JUGG+zCAua3QFAc3WkA2y4tEgbUjxpFqfunjoOykdcqke
dKkVs4j/uzdFg49Ftmb4OfvRH73oMSsh3EyYResBvJG09qnoWhpNFpo7atLwlcWm
g6H5Eov0H6SDBaFzLFT5gty8sOSd6I3wbU0p5zkCgYAwQe8+M7Su2v3mA0vbxbGb
W96El5n15YRa6JHOigC+5mBhXilnDE8qomFkfELDOnQ+hdkgqbFd7P1/+K5raV+I
aGh+dZd2MKnLevVoMexu40NQLVyJTOqumG05NNgmfg7VE8QUbXKfz+9pmfYFSZGS
Wx4EqMDhdG9wlTsHGb1I/Q==
-----END PRIVATE KEY-----
`)

var testCertPEM = []byte(`-----BEGIN CERTIFICATE-----
MIIDfDCCAmQCCQCAHkBfX03BnTANBgkqhkiG9w0BAQsFADB/MQswCQYDVQQGEwJD
TjELMAkGA1UECAwCQkoxEDAOBgNVBAcMB0JlaWppbmcxDjAMBgNVBAoMBUhTTEFN
MQwwCgYDVQQLDANSJkQxEjAQBgNVBAMMCWhzbGFtLmNvbTEfMB0GCSqGSIb3DQEJ
ARYQNzkxODc0MTU4QHFxLmNvbTAgFw0yMDA5MjMwMzE3NTdaGA8yMTIwMDgzMDAz
MTc1N1owfzELMAkGA1UEBhMCQ04xCzAJBgNVBAgMAkJKMRAwDgYDVQQHDAdCZWlq
aW5nMQ4wDAYDVQQKDAVIU0xBTTEMMAoGA1UECwwDUiZEMRIwEAYDVQQDDAloc2xh
bS5jb20xHzAdBgkqhkiG9w0BCQEWEDc5MTg3NDE1OEBxcS5jb20wggEiMA0GCSqG
SIb3DQEBAQUAA4IBDwAwggEKAoIBAQDa3lfVcQZg3Ra2DCeTPM9I8cv35Y+R4niX
J7c2U9TvGE3l8zfsLBXtdN4bSlmaimOnOmfx0aVJ8XwLqcIMspJmzG9UlGdlOfir
MTYCybvwhEf9bZc9lmLv27C4++4IljNF9sSv/Lnbdl5VNr+lY1xKRR3HPpwwuJj3
jh3TznzAnb0QnIRTyGGVThyE6uUQAgx8/taGenJDkzb7pry4kRvz+GgjAvhi/KOg
xho7G6PLfzXeS+iPyaMg5npd3B90XIzaaXr4/yffC5BUbynHhZLmWKJXSp7brjiZ
zpFV8np7wKYrtqXW4My2MtMASnvfXrCfTwQ3FU0biBsk7dQHCcuzAgMBAAEwDQYJ
KoZIhvcNAQELBQADggEBAA4rrtWczvjVpttxJ7pbXQlmvVrakPwqqKEQ09hxcoqY
EKkCucjJwFFQi1fNQBKpb+3BwlHIcfqdwpURiTwQjPmRgVhqdFqHE5pNF9EXdNm7
zaylUiu+ySKKHHnCVagM7UszovCoRYY3hq75UsGwR+9WWxOoWRz43NdOTBBDE9y7
JkRowySk9JE5isec+G0tDf6Fyj/3zWshWQalEH/Aq1Af0BMtWQL4VYXbealqK6rq
MOwPd7m67gCJlNREX2JnMDBM2A9QcAIzhYrHBx5w6UhUwSL6IFhJzdFXl4klsKUQ
cmw7rbPxsuPIyPlCobdtFoVpFN5vnOnF42nCb8tr0Xs=
-----END CERTIFICATE-----
`)
