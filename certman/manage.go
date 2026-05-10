package certman

import (
	"bytes"
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/libdns/libdns"
	"golang.org/x/crypto/acme"
	"golang.org/x/net/idna"

	"github.com/rehiy/libgo/logman"
)

type Manager struct {
	Email                  string
	DirectoryURL           string
	ExternalAccountBinding *acme.ExternalAccountBinding

	Cache  Cache
	Logger *logman.Logger

	DnsProvider interface {
		libdns.RecordAppender
		libdns.RecordDeleter
	}

	client   *acme.Client
	clientMu sync.Mutex

	state   map[certKey]*certState
	stateMu sync.Mutex
}

// GetCertificate 获取单个域名证书
func (m *Manager) GetCertificate(name string) (*tls.Certificate, error) {
	certs, err := m.GetCertificates([]string{name})
	if err != nil {
		return nil, err
	}
	return certs[0], nil
}

// GetCertificates 获取多域名证书（支持 SAN 和通配符）
// 所有域名将包含在同一张证书中
func (m *Manager) GetCertificates(names []string) ([]*tls.Certificate, error) {
	if len(names) == 0 {
		return nil, errors.New("missing domain names")
	}

	// 转换所有域名为 ASCII（正确处理通配符）
	asciiNames := make([]string, len(names))
	for i, name := range names {
		asciiName, err := toASCII(name)
		if err != nil {
			return nil, errors.New("domain name contains invalid character: " + name)
		}
		asciiNames[i] = asciiName
	}

	// 使用第一个域名作为主键
	ck := certKey{domain: asciiNames[0]}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	// 尝试从内存缓存加载
	cert, err := m.loadCert(ctx, ck)
	if err == nil {
		// 验证证书是否包含所有请求的域名
		if m.certContainsNames(cert, asciiNames) {
			return []*tls.Certificate{cert}, nil
		}
	}
	if err != nil && err != ErrCacheMiss {
		return nil, err
	}

	// 申请新证书
	cert, err = m.createCertWithNames(ctx, ck, asciiNames)
	if err != nil {
		return nil, err
	}

	// 异步保存到缓存
	go m.pemSave(context.Background(), ck, cert)

	return []*tls.Certificate{cert}, nil
}

// certContainsNames 检查证书是否包含所有请求的域名
func (m *Manager) certContainsNames(cert *tls.Certificate, names []string) bool {
	if cert.Leaf == nil {
		return false
	}
	for _, name := range names {
		if err := cert.Leaf.VerifyHostname(name); err != nil {
			return false
		}
	}
	return true
}

// toASCII 转换域名为 ASCII，正确处理通配符
func toASCII(name string) (string, error) {
	// 处理通配符域名
	if strings.HasPrefix(name, "*.") {
		ascii, err := idna.Lookup.ToASCII(name[2:])
		if err != nil {
			return "", err
		}
		return "*." + ascii, nil
	}
	return idna.Lookup.ToASCII(name)
}

func (m *Manager) loadCert(ctx context.Context, ck certKey) (*tls.Certificate, error) {
	m.stateMu.Lock()

	if s, ok := m.state[ck]; ok {
		m.stateMu.Unlock()
		s.RLock()
		defer s.RUnlock()
		return s.tlscert()
	}
	defer m.stateMu.Unlock()

	cert, err := m.pemLoad(ctx, ck)
	if err != nil {
		return nil, err
	}

	signer, ok := cert.PrivateKey.(crypto.Signer)
	if !ok {
		return nil, errors.New("private key cannot sign")
	}
	if m.state == nil {
		m.state = make(map[certKey]*certState)
	}

	s := &certState{
		key:  signer,
		cert: cert.Certificate,
		leaf: cert.Leaf,
	}
	m.state[ck] = s

	return cert, nil
}

func (m *Manager) createCert(ctx context.Context, ck certKey) (*tls.Certificate, error) {
	return m.createCertWithNames(ctx, ck, []string{ck.domain})
}

func (m *Manager) createCertWithNames(ctx context.Context, ck certKey, names []string) (*tls.Certificate, error) {
	state, err := m.certState(ck)
	if err != nil {
		return nil, err
	}

	if !state.locked {
		state.RLock()
		defer state.RUnlock()
		return state.tlscert()
	}

	defer state.Unlock()
	state.locked = false

	der, leaf, err := m.authorizedCert(ctx, state.key, ck, names)
	if err != nil {
		time.AfterFunc(time.Minute, func() {
			m.stateMu.Lock()
			defer m.stateMu.Unlock()
			s, ok := m.state[ck]
			if !ok {
				return
			}
			if _, err := validCertificate(ck, s.cert, s.key, time.Now()); err == nil {
				return
			}
			delete(m.state, ck)
		})
		return nil, err
	}

	state.cert = der
	state.leaf = leaf

	return state.tlscert()
}

func (m *Manager) certState(ck certKey) (*certState, error) {
	m.stateMu.Lock()
	defer m.stateMu.Unlock()

	if m.state == nil {
		m.state = make(map[certKey]*certState)
	}
	if state, ok := m.state[ck]; ok {
		return state, nil
	}

	// new locked state
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}

	state := &certState{
		key:    key,
		locked: true,
	}
	state.Lock() // will be unlocked by m.certState caller

	m.state[ck] = state

	return state, nil
}

func (m *Manager) authorizedCert(ctx context.Context, key crypto.Signer, ck certKey, names []string) (der [][]byte, leaf *x509.Certificate, err error) {
	req := &x509.CertificateRequest{
		Subject:  pkix.Name{CommonName: names[0]},
		DNSNames: names,
	}
	csr, err := x509.CreateCertificateRequest(rand.Reader, req, key)
	if err != nil {
		return nil, nil, err
	}

	m.Logger.Info("create client", "domain", names[0])
	client, err := m.acmeClient(ctx)
	if err != nil {
		return nil, nil, err
	}

	m.Logger.Info("authorize order", "domains", names)
	order, err := m.authorizedOrder(ctx, ck, names)
	if err != nil {
		return nil, nil, err
	}

	m.Logger.Info("finalize order", "domain", names[0])
	chain, _, err := client.CreateOrderCert(ctx, order.FinalizeURL, csr, true)
	if err != nil {
		return nil, nil, err
	}

	m.Logger.Info("verify certificate", "domain", names[0])
	leaf, err = validCertificate(ck, chain, key, time.Now())
	if err != nil {
		return nil, nil, err
	}

	return chain, leaf, nil
}

func (m *Manager) authorizedOrder(ctx context.Context, ck certKey, names []string) (*acme.Order, error) {
	order, err := m.client.AuthorizeOrder(ctx, acme.DomainIDs(names...))
	if err != nil {
		return nil, err
	}

	defer func() { go m.revokePendingAuthz(order.AuthzURLs) }()

	if order.Status == acme.StatusReady {
		return order, nil
	}
	if order.Status != acme.StatusPending {
		return nil, errors.New("invalid new order status " + order.Status)
	}

	for _, zurl := range order.AuthzURLs {
		m.Logger.Info("authorizing domain", "domains", names, "authz_url", zurl)

		authz, err := m.client.GetAuthorization(ctx, zurl)
		if err != nil {
			return nil, err
		}
		if authz.Status != acme.StatusPending {
			continue
		}

		var chal *acme.Challenge
		for _, c := range authz.Challenges {
			if c.Type == "dns-01" {
				chal = c
				break
			}
		}
		if chal == nil {
			return nil, errors.New("no viable challenge type found")
		}

		// 对于通配符域名，authz.Identifier.Value 不带 *. 前缀
		domain := authz.Identifier.Value
		if cleanup, err := m.fulfill(ctx, chal, domain); err != nil {
			return nil, err
		} else {
			defer cleanup()
		}

		if _, err := m.client.Accept(ctx, chal); err != nil {
			return nil, err
		}

		m.Logger.Info("waiting for authorization to be valid", "uri", authz.URI)
		if _, err := m.client.WaitAuthorization(ctx, authz.URI); err != nil {
			return nil, err
		}
	}

	return m.client.WaitOrder(ctx, order.URI)
}

func (m *Manager) revokePendingAuthz(uri []string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	for _, u := range uri {
		authz, err := m.client.GetAuthorization(ctx, u)
		if err == nil && authz.Status == acme.StatusPending {
			m.client.RevokeAuthorization(ctx, u)
		}
	}
}

func (m *Manager) fulfill(ctx context.Context, chal *acme.Challenge, domain string) (func(), error) {
	value, err := m.client.DNS01ChallengeRecord(chal.Token)
	if err != nil {
		return nil, err
	}

	record := []libdns.Record{
		&libdns.TXT{Name: "_acme-challenge", Text: value},
	}

	if _, err = m.DnsProvider.AppendRecords(ctx, domain, record); err != nil {
		return nil, err
	}

	// 等待 DNS 生效
	select {
	case <-time.After(30 * time.Second):
	case <-ctx.Done():
		// 上下文取消时清理 DNS 记录
		go m.DnsProvider.DeleteRecords(context.Background(), domain, record)
		return nil, ctx.Err()
	}

	return func() {
		go m.DnsProvider.DeleteRecords(context.Background(), domain, record)
	}, nil
}

func (m *Manager) accountKey(ctx context.Context) (crypto.Signer, error) {
	const keyName = "account.key"

	genKey := func() (*ecdsa.PrivateKey, error) {
		return ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	}

	if m.Cache == nil {
		return genKey()
	}

	data, err := m.Cache.Get(ctx, keyName)
	if err != nil {
		if err == ErrCacheMiss {
			key, err := genKey()
			if err != nil {
				return nil, err
			}
			var buf bytes.Buffer
			if err := encodeECDSAKey(&buf, key); err != nil {
				return nil, err
			}
			if err := m.Cache.Put(ctx, keyName, buf.Bytes()); err != nil {
				return nil, err
			}
			return key, nil
		}
		return nil, err
	}

	priv, _ := pem.Decode(data)
	if priv == nil || !strings.Contains(priv.Type, "PRIVATE") {
		return nil, errors.New("invalid account key found in cache")
	}

	return parsePrivateKey(priv.Bytes)
}

func (m *Manager) acmeClient(ctx context.Context) (*acme.Client, error) {
	m.clientMu.Lock()
	defer m.clientMu.Unlock()

	if m.client != nil {
		return m.client, nil
	}

	accountKey, err := m.accountKey(ctx)
	if err != nil {
		return nil, err
	}

	client := &acme.Client{
		DirectoryURL: m.DirectoryURL,
		UserAgent:    "autocert",
		Key:          accountKey,
	}

	account := &acme.Account{
		Contact:                []string{"mailto:" + m.Email},
		ExternalAccountBinding: m.ExternalAccountBinding,
	}

	_, err = client.Register(ctx, account, acme.AcceptTOS)
	if ae, ok := err.(*acme.Error); err == nil ||
		err == acme.ErrAccountAlreadyExists || (ok && ae.StatusCode == http.StatusConflict) {
		m.client = client
		err = nil
	}

	return m.client, err
}
