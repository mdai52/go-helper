package certify

import (
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"

	"golang.org/x/crypto/acme"
	"golang.org/x/crypto/acme/autocert"
)

// getClient 获取 ACME 客户端（单例）
func (m *Manager) getClient(ctx context.Context) (*acme.Client, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.client != nil {
		return m.client, nil
	}

	key, err := m.accountKey(ctx)
	if err != nil {
		return nil, err
	}

	client := &acme.Client{
		Key:          key,
		DirectoryURL: autocert.DefaultACMEDirectory,
	}
	if m.Manager.Client != nil && m.Manager.Client.DirectoryURL != "" {
		client.DirectoryURL = m.Manager.Client.DirectoryURL
	}

	// 注册账户
	_, err = client.Register(ctx, &acme.Account{
		Contact:                []string{"mailto:" + m.Manager.Email},
		ExternalAccountBinding: m.Manager.ExternalAccountBinding,
	}, autocert.AcceptTOS)
	if err != nil && !isAccountExists(err) {
		return nil, err
	}

	m.client = client
	return client, nil
}

// createCert 创建证书（从订单）
func (m *Manager) createCert(ctx context.Context, client *acme.Client, order *acme.Order, domains []string) (*tls.Certificate, error) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}

	csr, err := x509.CreateCertificateRequest(rand.Reader, &x509.CertificateRequest{
		Subject:  pkix.Name{CommonName: domains[0]},
		DNSNames: domains,
	}, key)
	if err != nil {
		return nil, err
	}

	der, _, err := client.CreateOrderCert(ctx, order.FinalizeURL, csr, true)
	if err != nil {
		return nil, err
	}

	cert := &tls.Certificate{
		PrivateKey:  key,
		Certificate: der,
	}
	if cert.Leaf, err = x509.ParseCertificate(der[0]); err != nil {
		return nil, err
	}

	// 异步保存到缓存
	if m.Cache != nil {
		go m.saveCert(context.Background(), domains, cert)
	}

	return cert, nil
}

// createCertFromAuthz 从授权创建证书（单域名）
func (m *Manager) createCertFromAuthz(ctx context.Context, client *acme.Client, domain string) (*tls.Certificate, error) {
	order, err := client.AuthorizeOrder(ctx, acme.DomainIDs(domain))
	if err != nil {
		return nil, err
	}
	order, err = client.WaitOrder(ctx, order.URI)
	if err != nil {
		return nil, err
	}
	return m.createCert(ctx, client, order, []string{domain})
}

// ==================== 辅助函数 ====================

// accountKey 获取或创建账户密钥
func (m *Manager) accountKey(ctx context.Context) (crypto.Signer, error) {
	if m.Cache == nil {
		return ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	}

	data, err := m.Cache.Get(ctx, "account.key")
	if err == autocert.ErrCacheMiss {
		key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		if err != nil {
			return nil, err
		}
		buf, err := keyToPEM(key)
		if err != nil {
			return nil, err
		}
		m.Cache.Put(ctx, "account.key", buf)
		return key, nil
	}
	if err != nil {
		return nil, err
	}
	return keyFromPEM(data)
}

// loadCert 从缓存加载证书
func (m *Manager) loadCert(domain string) (*tls.Certificate, error) {
	data, err := m.Cache.Get(context.Background(), domain+".crt")
	if err != nil {
		return nil, err
	}
	return tlsCertFromPEM(data)
}

// saveCert 保存证书到缓存（为所有域名保存）
func (m *Manager) saveCert(ctx context.Context, domains []string, cert *tls.Certificate) {
	data := certToPEM(cert)
	for _, domain := range domains {
		m.Cache.Put(ctx, domain+".crt", data)
	}
}
