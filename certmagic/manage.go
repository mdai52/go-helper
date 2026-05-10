package certmagic

import (
	"context"
	"errors"
	"strings"
	"sync"

	"github.com/caddyserver/certmagic"
	"github.com/forgoer/openssl"
)

var (
	magicPool   = map[string]*certmagic.Config{}
	magicPoolMu sync.RWMutex
)

func Manage(rq *RequestParam) error {
	skey := openssl.Md5ToString(rq.Email + rq.SecretKey + rq.CaType)

	magicPoolMu.Lock()
	defer magicPoolMu.Unlock()

	magic, ok := magicPool[skey]

	if !ok {
		magic = newMagic(newIssuer(rq), rq.StoragePath)
		magicPool[skey] = magic
	}

	magicPool[rq.Domain] = magic

	domains := strings.Split(rq.Domain, ",")
	return magic.ManageAsync(context.Background(), domains)
}

func Unmanage(domain string) {
	magicPoolMu.Lock()
	defer magicPoolMu.Unlock()

	magic, ok := magicPool[domain]
	domains := strings.Split(domain, ",")

	if ok {
		delete(magicPool, domain)
		for _, d := range domains {
			magic.RevokeCert(context.Background(), d, 0, false)
		}
	}
}

func CertDetail(domain string) (*Certificate, error) {
	magicPoolMu.RLock()
	magic, ok := magicPool[domain]
	magicPoolMu.RUnlock()

	if !ok {
		return nil, errors.New("not exist or deleted")
	}

	crt, err := magic.CacheManagedCertificate(context.Background(), domain)
	if err != nil {
		return nil, err
	}

	pk, err := certmagic.PEMEncodePrivateKey(crt.Certificate.PrivateKey)
	if err != nil {
		return nil, err
	}

	cert := &Certificate{
		Names:       crt.Names,
		NotAfter:    crt.Leaf.NotAfter.Unix(),
		NotBefore:   crt.Leaf.NotBefore.Unix(),
		Certificate: crt.Certificate.Certificate,
		PrivateKey:  pk,
	}

	// 安全访问 Issuer 字段
	if len(crt.Leaf.Issuer.Organization) > 0 {
		cert.Issuer = map[string]any{
			"Organization": crt.Leaf.Issuer.Organization[0],
		}
		if len(crt.Leaf.Issuer.Country) > 0 {
			cert.Issuer["Country"] = crt.Leaf.Issuer.Country[0]
		}
	}
	if crt.Leaf.Issuer.CommonName != "" {
		if cert.Issuer == nil {
			cert.Issuer = make(map[string]any)
		}
		cert.Issuer["CommonName"] = crt.Leaf.Issuer.CommonName
	}

	return cert, nil
}
