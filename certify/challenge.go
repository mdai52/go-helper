package certify

import (
	"context"
	"crypto/tls"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/libdns/libdns"
	"golang.org/x/crypto/acme"
)

// 错误定义
var (
	ErrNoDomains      = errors.New("no domains specified")
	ErrNoDNSProvider  = errors.New("DNS provider required for DNS-01 challenge")
	ErrNoDNSChallenge = errors.New("no dns-01 challenge")
	ErrNoChallenge    = errors.New("no supported challenge")
)

// getSingleCert 获取单域名证书（HTTP-01/TLS-ALPN-01）
func (m *Manager) getSingleCert(domain string) (*tls.Certificate, error) {
	// 尝试从缓存加载
	if m.Cache != nil {
		if cert, err := m.loadCert(domain); err == nil && !isCertExpired(cert) {
			m.emitEvent("cached", map[string]any{"domains": []string{domain}})
			return cert, nil
		}
	}

	m.emitEvent("obtaining", map[string]any{"domains": []string{domain}})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	client, err := m.getClient(ctx)
	if err != nil {
		m.emitEvent("failed", map[string]any{"domains": []string{domain}, "error": err.Error()})
		return nil, err
	}

	// 获取授权
	authz, err := client.Authorize(ctx, domain)
	if err != nil {
		m.emitEvent("failed", map[string]any{"domains": []string{domain}, "error": err.Error()})
		return nil, err
	}

	// 完成验证
	if authz.Status == acme.StatusPending {
		if err := m.solveChallenge(ctx, client, authz); err != nil {
			m.emitEvent("failed", map[string]any{"domains": []string{domain}, "error": err.Error()})
			return nil, err
		}
	}

	cert, err := m.createCertFromAuthz(ctx, client, domain)
	if err != nil {
		m.emitEvent("failed", map[string]any{"domains": []string{domain}, "error": err.Error()})
		return nil, err
	}

	m.emitEvent("obtained", map[string]any{"domains": []string{domain}, "expires": cert.Leaf.NotAfter.Unix()})
	return cert, nil
}

// getCertDNS01 使用 DNS-01 获取证书（支持通配符和多域名）
func (m *Manager) getCertDNS01(domains ...string) (*tls.Certificate, error) {
	if m.DNSProvider == nil {
		return nil, ErrNoDNSProvider
	}

	// 尝试从缓存加载
	if m.Cache != nil {
		for _, domain := range domains {
			if cert, err := m.loadCert(domain); err == nil {
				if !isCertExpired(cert) && certContainsAll(cert, domains) {
					m.emitEvent("cached", map[string]any{"domains": domains})
					return cert, nil
				}
			}
		}
	}

	m.emitEvent("obtaining", map[string]any{"domains": domains})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	client, err := m.getClient(ctx)
	if err != nil {
		m.emitEvent("failed", map[string]any{"domains": domains, "error": err.Error()})
		return nil, err
	}

	// 创建订单
	order, err := client.AuthorizeOrder(ctx, acme.DomainIDs(domains...))
	if err != nil {
		return nil, err
	}

	// 收集挑战和 DNS 记录（按 zone 分组）
	var challenges []*acme.Challenge
	zoneRecords := make(map[string][]libdns.Record)

	for _, authzURL := range order.AuthzURLs {
		authz, err := client.GetAuthorization(ctx, authzURL)
		if err != nil {
			return nil, err
		}
		if authz.Status != acme.StatusPending {
			continue
		}

		chal := findChallenge(authz, "dns-01")
		if chal == nil {
			return nil, ErrNoDNSChallenge
		}
		challenges = append(challenges, chal)

		value, err := client.DNS01ChallengeRecord(chal.Token)
		if err != nil {
			return nil, err
		}

		// 处理通配符域名
		challengeDomain := authz.Identifier.Value
		if isWildcard(challengeDomain) {
			challengeDomain = challengeDomain[2:]
		}

		zone, recordName := extractZoneAndRecordName(challengeDomain)
		zoneRecords[zone] = append(zoneRecords[zone], &libdns.TXT{Name: recordName, Text: value})
	}

	// 确保清理 DNS 记录
	defer func() {
		for zone, records := range zoneRecords {
			go m.DNSProvider.DeleteRecords(context.Background(), zone, records)
		}
	}()

	// 添加所有 DNS 记录
	for zone, records := range zoneRecords {
		m.logInfo("adding DNS records", "zone", zone, "count", len(records))
		m.DNSProvider.DeleteRecords(ctx, zone, records)
		if _, err := m.DNSProvider.AppendRecords(ctx, zone, records); err != nil {
			return nil, err
		}
	}

	// 等待 DNS 生效
	m.logInfo("waiting for DNS propagation")
	time.Sleep(30 * time.Second)

	// 接受所有挑战
	for _, chal := range challenges {
		if _, err := client.Accept(ctx, chal); err != nil {
			return nil, err
		}
	}

	// 等待订单就绪并创建证书
	order, err = client.WaitOrder(ctx, order.URI)
	if err != nil {
		return nil, err
	}

	m.logInfo("order finalized", "domains", domains)

	cert, err := m.createCert(ctx, client, order, domains)
	if err != nil {
		m.emitEvent("failed", map[string]any{"domains": domains, "error": err.Error()})
		return nil, err
	}

	m.logInfo("certificate obtained", "domain", domains[0], "expires", cert.Leaf.NotAfter.Format("2006-01-02"))
	m.emitEvent("obtained", map[string]any{"domains": domains, "expires": cert.Leaf.NotAfter.Unix()})
	return cert, nil
}

// solveChallenge 完成 HTTP-01 或 TLS-ALPN-01 验证
func (m *Manager) solveChallenge(ctx context.Context, client *acme.Client, authz *acme.Authorization) error {
	chal := findChallenge(authz, "http-01")
	if chal == nil {
		chal = findChallenge(authz, "tls-alpn-01")
	}
	if chal == nil {
		return ErrNoChallenge
	}

	m.logInfo("accepting challenge", "type", chal.Type, "domain", authz.Identifier.Value)
	if _, err := client.Accept(ctx, chal); err != nil {
		return err
	}

	_, err := client.WaitAuthorization(ctx, authz.URI)
	return err
}

// ==================== 辅助函数 ====================

// findChallenge 查找指定类型的挑战
func findChallenge(authz *acme.Authorization, challengeType string) *acme.Challenge {
	for _, c := range authz.Challenges {
		if c.Type == challengeType {
			return c
		}
	}
	return nil
}

// extractZoneAndRecordName 提取 zone 和 DNS 记录名
// 例如: example.com -> zone="example.com", name="_acme-challenge"
//       www.example.com -> zone="example.com", name="_acme-challenge.www"
func extractZoneAndRecordName(domain string) (zone, recordName string) {
	parts := strings.Split(domain, ".")
	if len(parts) <= 2 {
		return domain, "_acme-challenge"
	}
	zone = strings.Join(parts[len(parts)-2:], ".")
	recordName = "_acme-challenge." + strings.Join(parts[:len(parts)-2], ".")
	return
}

// isAccountExists 检查账户是否已存在
func isAccountExists(err error) bool {
	if err == acme.ErrAccountAlreadyExists {
		return true
	}
	if ae, ok := err.(*acme.Error); ok && ae.StatusCode == http.StatusConflict {
		return true
	}
	return false
}

// certContainsAll 检查证书是否包含所有域名
func certContainsAll(cert *tls.Certificate, domains []string) bool {
	if cert.Leaf == nil {
		return false
	}
	for _, d := range domains {
		if cert.Leaf.VerifyHostname(d) != nil {
			return false
		}
	}
	return true
}

// isCertExpired 检查证书是否过期（提前 30 天续期）
func isCertExpired(cert *tls.Certificate) bool {
	return cert.Leaf == nil || time.Now().Add(30*24*time.Hour).After(cert.Leaf.NotAfter)
}

// logInfo 安全日志
func (m *Manager) logInfo(msg string, args ...any) {
	if m.Logger != nil {
		m.Logger.Info(msg, args...)
	}
}