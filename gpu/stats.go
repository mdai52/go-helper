package gpu

import (
	"context"
	"errors"
	"strings"

	"github.com/jaypipes/ghw"
	ghwgpu "github.com/jaypipes/ghw/pkg/gpu"
)

// DeviceStat GPU 原始采集结果
// Address 为规范化后的 PCI 地址，供上层决定是否生成稳定设备标识
// Vendor 使用统一的小写厂商名：nvidia | amd | intel | apple
// 该结构只承载采集结果，不包含展示序号等业务字段
type DeviceStat struct {
	Address     string  `json:"address"`
	Name        string  `json:"name"`
	Vendor      string  `json:"vendor"`
	MemoryUsed  uint64  `json:"memoryUsed"`
	MemoryTotal uint64  `json:"memoryTotal"`
	Utilization float64 `json:"utilization"`
	Temperature int     `json:"temperature"`
	PowerUsage  float64 `json:"powerUsage"`
	FanSpeed    int     `json:"fanSpeed"`
}

// GetGPUStats 采集所有 GPU 原始统计信息
// Linux/Windows：ghw 发现 → 按厂商分发采集（nvidia/amd/intel）
// macOS：通过 ioreg 采集 Apple GPU（ghw 不支持 macOS GPU 发现）
func GetGPUStats(ctx context.Context) ([]*DeviceStat, error) {
	var nvidiaCards, amdCards, intelCards []*ghwgpu.GraphicsCard
	var collectErr error

	gpuInfo, err := ghw.GPU()
	if err == nil && gpuInfo != nil {
		for _, card := range gpuInfo.GraphicsCards {
			if card.DeviceInfo == nil || card.DeviceInfo.Vendor == nil {
				continue
			}
			vid := strings.ToLower(card.DeviceInfo.Vendor.ID)

			if virtualVendorIDs[vid] {
				continue
			}
			if isIntegratedGPU(card) {
				continue
			}

			switch vid {
			case "10de":
				nvidiaCards = append(nvidiaCards, card)
			case "1002":
				amdCards = append(amdCards, card)
			case "8086":
				intelCards = append(intelCards, card)
			}
		}
	} else if err != nil {
		collectErr = err
	}

	var result []*DeviceStat

	nvidiaStats, err := collectNvidiaCards(ctx, nvidiaCards)
	if err != nil {
		collectErr = errors.Join(collectErr, err)
	}
	result = append(result, nvidiaStats...)

	amdStats, err := collectAMDCards(ctx, amdCards)
	if err != nil {
		collectErr = errors.Join(collectErr, err)
	}
	result = append(result, amdStats...)

	result = append(result, collectIntelCards(intelCards)...)

	appleStats, err := collectAppleGPUs(ctx)
	if err != nil {
		collectErr = errors.Join(collectErr, err)
	}
	result = append(result, appleStats...)

	return result, collectErr
}

func isIntegratedGPU(card *ghwgpu.GraphicsCard) bool {
	if card.DeviceInfo == nil || card.DeviceInfo.Vendor == nil {
		return false
	}
	// 只对 Intel 做核显过滤；AMD APU 核显通常无独立 VRAM，后续采集时自然跳过
	if strings.ToLower(card.DeviceInfo.Vendor.ID) != "8086" {
		return false
	}
	name := ""
	if card.DeviceInfo.Product != nil {
		name = card.DeviceInfo.Product.Name
	}
	// Intel Arc 独显保留
	if strings.Contains(name, "Arc") {
		return false
	}
	for _, kw := range IGPUKeywords {
		if strings.Contains(name, kw) {
			return true
		}
	}
	return false
}

func cardProductName(card *ghwgpu.GraphicsCard) string {
	if card.DeviceInfo != nil && card.DeviceInfo.Product != nil {
		return card.DeviceInfo.Product.Name
	}
	return ""
}

func cardAddress(card *ghwgpu.GraphicsCard) string {
	if card == nil {
		return ""
	}
	return normalizePCIAddress(card.Address)
}
