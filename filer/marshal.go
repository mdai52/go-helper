package filer

import (
	"encoding/json"

	"gopkg.in/yaml.v3"
)

// JSON 格式保存数据

func SaveJson(file string, data any) error {
	buf, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	// 写入文件
	if err := Write(file, buf); err != nil {
		return err
	}
	return nil
}

func ReadJson(file string, data any) error {
	bytes, err := Read(file)
	if err != nil {
		return err
	}
	// 解析 JSON 数据
	if err := json.Unmarshal(bytes, data); err != nil {
		return err
	}
	return nil
}

func ReadJsonT[T any](file string) (T, error) {
	var result T
	bytes, err := Read(file)
	if err != nil {
		return result, err
	}
	// 解析 JSON 数据
	if err := json.Unmarshal(bytes, &result); err != nil {
		return result, err
	}
	return result, nil
}

// YAML 格式保存数据

func SaveYaml(file string, data any) error {
	buf, err := yaml.Marshal(data)
	if err != nil {
		return err
	}
	// 写入文件
	if err := Write(file, buf); err != nil {
		return err
	}
	return nil
}

func ReadYaml(file string, data any) error {
	bytes, err := Read(file)
	if err != nil {
		return err
	}
	// 解析 YAML 数据
	if err := yaml.Unmarshal(bytes, data); err != nil {
		return err
	}
	return nil
}

func ReadYamlT[T any](file string) (T, error) {
	var result T
	bytes, err := Read(file)
	if err != nil {
		return result, err
	}
	// 解析 YAML 数据
	if err := yaml.Unmarshal(bytes, &result); err != nil {
		return result, err
	}
	return result, nil
}
