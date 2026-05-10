package tencent

type RequestParam struct {
	Service   string `note:"产品名称" json:"service"`
	Version   string `note:"接口版本" json:"version"`
	Action    string `note:"接口名称" json:"action"`
	Payload   any    `note:"结构化数据" json:"payload"`
	Region    string `note:"资源所在区域" json:"region"`
	Endpoint  string `note:"指定接口域名" json:"endpoint"`
	SecretId  string `note:"访问密钥 Id" json:"secretId"`
	SecretKey string `note:"访问密钥 Key" json:"secretKey"`
	Debug     bool   `note:"是否开启调试" json:"debug"`
}

type ResponseData struct {
	Response any
}
