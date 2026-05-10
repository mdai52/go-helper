package certmagic

type RequestParam struct {
	Email       string `note:"邮箱" json:"email"`
	Domain      string `note:"域名" json:"domain"`
	CaType      string `note:"Ca 类型" json:"caType"`
	Provider    string `note:"Ca 提供商" json:"provider"`
	SecretId    string `note:"密钥 ID" json:"secretId"`
	SecretKey   string `note:"密钥" json:"secretKey"`
	EabKeyId    string `note:"EAB密钥 ID" json:"eabKeyId"`
	EabMacKey   string `note:"EAB密钥" json:"eabMacKey"`
	StoragePath string `note:"存储目录" json:"storagePath"`
}

type Certificate struct {
	Names       []string `json:"names"`
	NotAfter    int64    `json:"notAfter"`
	NotBefore   int64    `json:"notBefore"`
	Certificate [][]byte `json:"certificate"`
	PrivateKey  []byte   `json:"privateKey"`
	Issuer      map[string]any `json:"issuer"`
}
