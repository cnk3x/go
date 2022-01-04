package httpx

import (
	"bytes"
	"crypto"
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"strings"
	"time"
)

var (
	ErrMissPrivateKey            = errors.New("certificate: private key miss")
	ErrMissPublicKey             = errors.New("certificate: public key miss")
	ErrNotEffective              = errors.New("certificate: is not effective yet")
	ErrExpired                   = errors.New("certificate: has been expired")
	ErrKeyPairTypeNotMatch       = errors.New("certificate: private key type does not match public key type")
	ErrKeyPairNotMatch           = errors.New("certificate: private key does not match public key")
	ErrUnknownPublicKeyAlgorithm = errors.New("certificate: unknown public key algorithm")
	ErrUnknownPrivateKeyType     = errors.New("certificate: unknown private key type")
)

func GetCertificateBytes(tlscert *tls.Certificate) ([]byte, error) {
	// contains PEM-encoded data
	var buf bytes.Buffer

	// private
	switch key := tlscert.PrivateKey.(type) {
	case *ecdsa.PrivateKey:
		b, err := x509.MarshalECPrivateKey(key)
		if err != nil {
			return nil, err
		}
		if err := pem.Encode(&buf, &pem.Block{
			Type:  "EC PRIVATE KEY",
			Bytes: b,
		}); err != nil {
			return nil, err
		}
	case *rsa.PrivateKey:
		if err := pem.Encode(&buf, &pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(key),
		}); err != nil {
			return nil, err
		}
	default:
		return nil, ErrUnknownPrivateKeyType
	}

	// public
	for _, b := range tlscert.Certificate {
		pb := &pem.Block{Type: "CERTIFICATE", Bytes: b}
		if err := pem.Encode(&buf, pb); err != nil {
			return nil, err
		}
	}

	return buf.Bytes(), nil
}

func GetCertificate(data []byte) (*tls.Certificate, error) {
	// private
	priv, pub := pem.Decode(data)
	if priv == nil || !strings.Contains(priv.Type, "PRIVATE") {
		return nil, ErrMissPrivateKey
	}
	privKey, err := parsePrivateKey(priv.Bytes)
	if err != nil {
		return nil, err
	}

	// public
	var pubDER [][]byte
	for len(pub) > 0 {
		var b *pem.Block
		b, pub = pem.Decode(pub)
		if b == nil {
			break
		}
		pubDER = append(pubDER, b.Bytes)
	}
	if len(pub) > 0 {
		// Leftover content not consumed by pem.Decode. Corrupt. Ignore.
		return nil, ErrMissPublicKey
	}

	// verify and create TLS cert
	leaf, err := validCert(pubDER, privKey, time.Now())
	if err != nil {
		return nil, err
	}
	tlscert := &tls.Certificate{
		Certificate: pubDER,
		PrivateKey:  privKey,
		Leaf:        leaf,
	}
	return tlscert, nil
}

// Attempt to parse the given private key DER block. OpenSSL 0.9.8 generates
// PKCS#1 private keys by default, while OpenSSL 1.0.0 generates PKCS#8 keys.
// OpenSSL ecparam generates SEC1 EC private keys for ECDSA. We try all three.
//
// Inspired by parsePrivateKey in crypto/tls/tls.go.
func parsePrivateKey(der []byte) (crypto.Signer, error) {
	if key, err := x509.ParsePKCS1PrivateKey(der); err == nil {
		return key, nil
	}
	if key, err := x509.ParsePKCS8PrivateKey(der); err == nil {
		switch key := key.(type) {
		case *rsa.PrivateKey:
			return key, nil
		case *ecdsa.PrivateKey:
			return key, nil
		default:
			return nil, errors.New("unknown private key type in PKCS#8 wrapping")
		}
	}
	if key, err := x509.ParseECPrivateKey(der); err == nil {
		return key, nil
	}
	return nil, errors.New("failed to parse private key")
}

// validCert parses a cert chain provided as der argument and verifies the leaf and der[0]
// correspond to the private key, the domain and key type match, and expiration dates
// are valid. It doesn't do any revocation checking.
//
// The returned value is the verified leaf cert.
func validCert(der [][]byte, key crypto.Signer, now time.Time) (leaf *x509.Certificate, err error) {
	var n int
	for _, b := range der {
		n += len(b)
	}
	pub := make([]byte, n)
	n = 0
	for _, b := range der {
		n += copy(pub[n:], b)
	}
	x509Cert, err := x509.ParseCertificates(pub)
	if err != nil || len(x509Cert) == 0 {
		return nil, ErrMissPublicKey
	}
	// verify the leaf is not expired and matches the domain name
	leaf = x509Cert[0]
	if now.Before(leaf.NotBefore) {
		return nil, ErrNotEffective
	}
	if now.After(leaf.NotAfter) {
		return nil, ErrExpired
	}
	// if err := leaf.VerifyHostname(domain); err != nil {
	// 	return nil, err
	// }
	// ensure the leaf corresponds to the private key and matches the certKey type
	switch pub := leaf.PublicKey.(type) {
	case *rsa.PublicKey:
		prv, ok := key.(*rsa.PrivateKey)
		if !ok {
			return nil, ErrKeyPairTypeNotMatch
		}
		if pub.N.Cmp(prv.N) != 0 {
			return nil, ErrKeyPairNotMatch
		}
	case *ecdsa.PublicKey:
		prv, ok := key.(*ecdsa.PrivateKey)
		if !ok {
			return nil, ErrKeyPairTypeNotMatch
		}
		if pub.X.Cmp(prv.X) != 0 || pub.Y.Cmp(prv.Y) != 0 {
			return nil, ErrKeyPairNotMatch
		}
	default:
		return nil, ErrUnknownPublicKeyAlgorithm
	}
	return leaf, nil
}
