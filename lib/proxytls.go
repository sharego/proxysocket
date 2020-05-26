package lib

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"time"
	"strings"
	"net"
)

// DefaultCa system built-in Root Certificate
// func DefaultCa() (cert tls.Certificate) {
// 	certPem := []byte(`-----BEGIN CERTIFICATE-----
// 	MIID8TCCAtmgAwIBAgIJAKZF6Lqx4mJNMA0GCSqGSIb3DQEBCwUAMIGOMQswCQYD
// 	VQQGEwJDTjEQMA4GA1UECAwHQmVpamluZzEQMA4GA1UEBwwHQmVpamluZzERMA8G
// 	A1UECgwIUGVyc29uYWwxEDAOBgNVBAsMB1hpYW93ZWkxGTAXBgNVBAMMEHhpYW93
// 	ZWkgcGVyc29uYWwxGzAZBgkqhkiG9w0BCQEWDHh3QHh3c2VhLmNvbTAeFw0yMDA1
// 	MjYxNDE3MDJaFw0zMDA1MjQxNDE3MDJaMIGOMQswCQYDVQQGEwJDTjEQMA4GA1UE
// 	CAwHQmVpamluZzEQMA4GA1UEBwwHQmVpamluZzERMA8GA1UECgwIUGVyc29uYWwx
// 	EDAOBgNVBAsMB1hpYW93ZWkxGTAXBgNVBAMMEHhpYW93ZWkgcGVyc29uYWwxGzAZ
// 	BgkqhkiG9w0BCQEWDHh3QHh3c2VhLmNvbTCCASIwDQYJKoZIhvcNAQEBBQADggEP
// 	ADCCAQoCggEBAM5HSADKscTViiM3g3VLA9lweiNMyWAhq8oykcuQe02yKQhzGhEP
// 	lsuQ3YM9M8iwlN6/5WjPj8FL6hc4ShEMTtheaD8fxg9vPa3/H964QmBz9x7U/goQ
// 	XBX6CxmH+vF9hC4CqNG1cjChp5yeeat9Kt9b/zj/PDIO2Yeg4RkdKjbO6fR+OK2X
// 	BqorfeBcSXfWt5/DTrkEIbX5iL0VRbV1QVzG0IaUkMA7427/kfoe+anz4zo2ILK+
// 	A09F2aVYX5TwoazhucFCg9SlXYFIIU1SNAnnMu4SCtKofwRBLrUaBsSQGhpFdbF6
// 	8Ldb1L0I1taAXdGWFdAiOYcKp98PscGTJFcCAwEAAaNQME4wHQYDVR0OBBYEFIL6
// 	RnOTR3YLHIDa4KzCdnt7qqcGMB8GA1UdIwQYMBaAFIL6RnOTR3YLHIDa4KzCdnt7
// 	qqcGMAwGA1UdEwQFMAMBAf8wDQYJKoZIhvcNAQELBQADggEBAE/pGuMuIHk3TNlN
// 	VrhcF5K+YVSVYfKHCjgdgVmy2Kt+6CqerbOFHrOjNRvbPqHb56yd09xCCLFH1Qxh
// 	iFx/vZUYEJ/YuYi6vpgvGbDxXqW2oOr70J6jl3dbO9aKCH7GNxbKLUp0HssT0RAZ
// 	PG99zEO61WqwZ/efm5JKmcnpUXFB0Lo7uv5m4mPe4bQPKtGRoJTP+9Y/pNsdtlUr
// 	nuzvdK7v5nDAHwsa3sfs18G5ogNZKMKxY5rFomkkE7Or0WidzyuotV6zLlie56VW
// 	Fwq+IFbYQK2ko0VELBxnI+SaZnpN+JUrqprNU5kyPlLONhaD818dpL62CKi3jXS1
// 	w3xIaFw=
// 	-----END CERTIFICATE-----`)
// 	keyPem := []byte(`-----BEGIN PRIVATE KEY-----
// 	MIIEwAIBADANBgkqhkiG9w0BAQEFAASCBKowggSmAgEAAoIBAQDOR0gAyrHE1Yoj
// 	N4N1SwPZcHojTMlgIavKMpHLkHtNsikIcxoRD5bLkN2DPTPIsJTev+Voz4/BS+oX
// 	OEoRDE7YXmg/H8YPbz2t/x/euEJgc/ce1P4KEFwV+gsZh/rxfYQuAqjRtXIwoaec
// 	nnmrfSrfW/84/zwyDtmHoOEZHSo2zun0fjitlwaqK33gXEl31refw065BCG1+Yi9
// 	FUW1dUFcxtCGlJDAO+Nu/5H6Hvmp8+M6NiCyvgNPRdmlWF+U8KGs4bnBQoPUpV2B
// 	SCFNUjQJ5zLuEgrSqH8EQS61GgbEkBoaRXWxevC3W9S9CNbWgF3RlhXQIjmHCqff
// 	D7HBkyRXAgMBAAECggEBAK6W3FV3OYD8r99gxA4JgOeP8IBiJGsN9KW9qXKfBg3b
// 	xikVqrFX/WysXyAOM/8fndDuoE/WpbiX9TjT9rR5M16kgR00WmGD6LOVJLdQQzX5
// 	0OYyphWEhTxAlxZz5ixw7Og4bgSYy15n5EKGSzqfRSMpbVojhSJlOS43N24XJjyd
// 	3PXZmtwMXnHiGT+dNWeSshicKKwNaM8ShD8uMa0wVpkVDjaZCvsy3St2vGD8Lxns
// 	6C5pMVeLkfKdTMWzQ6Xug11i0zpoUPvhGSQTLOKn+kNQWcSjrLjC8OILoMW6BtvW
// 	ZukGG/s8CJmRku/Q4F5tUaUbFYqII6fP2qEhVu2Mv0ECgYEA84GltM35XIXUMTNS
// 	wam66v426z5FqSCrgWnwugg4iwqWPgba9wBgqof59qkc8dNQo/j8NDPofZ7cmjAO
// 	s70/hnqBEA4Rk1cH4MIJ2KjXUivyBppWr5V8LtphoLFyoIdPPRREI6A/Bjq9OOHG
// 	2X6UFzkzPC9VCQQKBkICioVz/7cCgYEA2NypBYQLeecfY/crPo71XsPY6WlMWOSi
// 	QR4aWx5XcJMI5hKIdJKtz8yXVQKJsZk1nzno0nxF5luoGIPirVPmUpfoAn7ClXNW
// 	Nv4EhS0V7hPVJeE0MI0Rdq6YNGnP4+4UHZtBdmRTbAE1ZNwtU5mW3eEV+ysiVGUi
// 	nZNEbwnawGECgYEAqgvRcg+coZX7dlhG5GLo2w2nwGN+NftQiVE4AFGZWs+L18jl
// 	xDOJTc0jK7MxOVa+K6PGS5YsNv1nRG1m9vGjmP+XfWS1tVTDdZtLUPenVgTGQQIq
// 	ovca/4UVRChmyd9z4E10wAQ5vtnqJfzU5jFUrVsOgmrzURZ3AaO7Loy/UsUCgYEA
// 	k+OfbICJJv/j21NS7V2mnZKn2T2S8EgzEn9J7/u4G7BZ8DSRVBY3bC+UwEdzyWa3
// 	N7fRO8G+FYNKFjXSnutJdefcM99oKW03TVdbk8qUWwCxahyzb6y0TSBx8cR7HnBc
// 	zXf0Y521ekE0vWydiJaEWRnH2Lqota8mtRkaITMyN+ECgYEAuoMOgqM4szoeldk/
// 	Lkr0p61LAjsZl1OEBvQvmWyQVPq00dQ14o6PPA2c4CaT+u5RoCZFrpEkv0TwFunF
// 	By6gBFd1tyKfQd4tHHYufOPwGOan0YdFrxlcGOX08mCLSMByN+yN2Zqvdy3zHsL6
// 	oE3XDHsfuLpP75l3z0Yd8OfGReI=
// 	-----END PRIVATE KEY-----`)
// 	cert, err := tls.X509KeyPair(certPem, keyPem)
// 	if err != nil {
// 		log.Fatalf("Parse Default CA Error: %s", err)
// 	}
// 	return cert
// }


// DefaultCa make inner ca cert
func DefaultCa() tls.Certificate {
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		panic(err)
	}
	template := x509.Certificate{SerialNumber: big.NewInt(1)}
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		panic(err)
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		log.Fatalf("Parse Default CA Error: %s", err)
	}
	return tlsCert
}

// GenerateServerCert make a tls certificate
func GenerateServerCert(name string, ip string, ca tls.Certificate) (cert tls.Certificate, err error) {
	ip127 := net.ParseIP("127.0.0.1")
	ips := []net.IP{ip127}
	if len(strings.TrimSpace(ip)) > 0 {
		ipaddr := net.ParseIP(ip)
		if ipaddr != nil && !ip127.Equal(ipaddr) {
			ips = []net.IP{ipaddr, ip127}
		}
	}

	return GenerateIPCert(name, ips, ca)
}

// GenerateCert make a tls certificate
func GenerateCert(name string, ca tls.Certificate) (cert tls.Certificate, err error) {
	return GenerateIPCert(name, nil, ca)
}

// GenerateIPCert make a tls certificate
func GenerateIPCert(name string, ips []net.IP, ca tls.Certificate) (cert tls.Certificate, err error) {
	cer := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Country:            []string{"CN"},
			Organization:       []string{"Personal"},
			OrganizationalUnit: []string{"Xiaowei"},
			Province:           []string{"Beijing"},
			CommonName:         name,
			Locality:           []string{"Beijing"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(1, 0, 0),
		BasicConstraintsValid: true,
		IsCA:                  false,
		ExtKeyUsage: []x509.ExtKeyUsage{
			x509.ExtKeyUsageClientAuth,
			x509.ExtKeyUsageServerAuth,
		},
		KeyUsage:       x509.KeyUsageDigitalSignature | x509.KeyUsageDataEncipherment,
		EmailAddresses: []string{"xw@xwsea.com"},
	}

	if ips!=nil && len(ips) > 0 {
		cer.IPAddresses = ips
	}

	key, err := rsa.GenerateKey(rand.Reader, 2048)

	if err != nil {
		return
	}

	cacert, err := x509.ParseCertificate(ca.Certificate[0])
	if err != nil {
		return
	}

	certDER, err := x509.CreateCertificate(rand.Reader, cer, cacert, &key.PublicKey, ca.PrivateKey)
	if err != nil {
		return
	}

	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return
	}

	return tlsCert, nil
}
