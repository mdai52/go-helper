package alibaba

type RequestParam struct {
	Service   string `note:"产品名称" json:"service"`
	Version   string `note:"接口版本" json:"version"`
	Action    string `note:"接口名称" json:"action"`
	Query     any    `note:"结构化数据" json:"query"`
	Payload   any    `note:"结构化数据" json:"payload"`
	RegionId  string `note:"资源所在区域" json:"regionId"`
	Endpoint  string `note:"指定接口域名" json:"endpoint"`
	SecretId  string `note:"访问密钥 Id" json:"secretId"`
	SecretKey string `note:"访问密钥 Key" json:"secretKey"`
}

type ResponseData struct {
	Response any
}

//// Endpoint

type EndpointItem struct {
	Id        string
	Endpoint  string
	Namespace string
	Protocols struct {
		Protocols []string
	}
	SerivceCode string
	Type        string
}

type EndpointBody struct {
	Endpoints struct {
		Endpoint []EndpointItem
	}
	RequestId string
	Success   bool
}
