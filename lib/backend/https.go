package backend

import (
	"bytes"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"

	"golang.org/x/crypto/ocsp"
)

const (
	maxConn int = 40
)

type (
	HTTPDB struct {
		sni        string
		url        string
		CertPool   string
		ClientCert string

		client    *http.Client
		tlsconfig *tls.Config

		hpkp map[string]bool
	}
)

func NewDB(URL, CertPool, ClientCert string, HPKP []string) (*HTTPDB, error) {
	remote, err := url.Parse(URL)
	if err != nil {
		return nil, err
	}

	db := &HTTPDB{
		sni:        remote.Host,
		url:        URL,
		CertPool:   CertPool,
		ClientCert: ClientCert,
		client:     &http.Client{},
		tlsconfig: &tls.Config{
			ServerName:         remote.Host,
			MinVersion:         tls.VersionTLS12,
			MaxVersion:         tls.VersionTLS12,
			ClientSessionCache: tls.NewLRUClientSessionCache(maxConn),
			CurvePreferences: []tls.CurveID{
				tls.CurveP521,
				tls.CurveP384,
				tls.CurveP256,
			},
			CipherSuites: []uint16{
				tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			},
		},
		hpkp: make(map[string]bool),
	}

	for _, k := range HPKP {
		db.hpkp[k] = true
	}

	db.client.Transport = &http.Transport{
		MaxIdleConnsPerHost: maxConn,
		DialTLS:             db.DialerTLS,
	}

	return db, nil
}

func (db *HTTPDB) Request(q *Query) (*Reply, error) {
	req, err := json.Marshal(q)
	if err != nil {
		return nil, err
	}

	res, err := http.Post(db.url+q.Verb, "application/json", bytes.NewReader(req))
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status %d for %s", res.StatusCode, db.url+q.Verb)
	}

	buff := new(bytes.Buffer)
	io.Copy(buff, res.Body)

	resp := new(Reply)
	err = json.Unmarshal(buff.Bytes(), resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// TODO Security Issue : this code was audited 0 time
func (db *HTTPDB) DialerTLS(network, addr string) (conn net.Conn, err error) {
	var ocsprep *ocsp.Response
	certok := false
	hostok := false
	ocspok := false

	c, err := tls.Dial(network, addr, db.tlsconfig)
	if err != nil {
		return c, err
	}

	cstate := c.ConnectionState()

	if cstate.OCSPResponse != nil {
		ocsprep, err = ocsp.ParseResponse(cstate.OCSPResponse, nil)
		if err != nil {
			return nil, err
		}

		switch ocsprep.Status {
		case ocsp.Good, ocsp.Unknown:

		default:
			return nil, fmt.Errorf("invalid OCSP")
		}
	}

	for _, peercert := range cstate.PeerCertificates {
		der, err := x509.MarshalPKIXPublicKey(peercert.PublicKey)
		if err != nil {
			return nil, err
		}

		if !hostok && peercert.VerifyHostname(db.sni) == nil {
			hostok = true
		}

		if ocsprep != nil && !ocspok && ocsprep.CheckSignatureFrom(peercert) == nil {
			ocspok = true
		}

		rawhash := sha256.Sum256(der)
		hash := base64.StdEncoding.EncodeToString(rawhash[:])

		if valid, ok := db.hpkp[hash]; !certok && ok && valid {
			certok = true
		}
	}

	if len(db.hpkp) > 0 && !certok {
		return nil, fmt.Errorf("invalid HPKP")
	}

	if !hostok {
		return nil, fmt.Errorf("invalid SNI")
	}

	if ocsprep != nil && !ocspok {
		return nil, fmt.Errorf("invalid OCSP")
	}

	return c, nil
}
