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
	return Write(file, buf)
}

func ReadJson(file string, data any) error {
	bytes, err := Read(file)
	if err != nil {
		return err
	}
	return json.Unmarshal(bytes, data)
}

func ReadJsonT[T any](file string) (T, error) {
	var result T
	bytes, err := Read(file)
	if err != nil {
		return result, err
	}
	err = json.Unmarshal(bytes, &result)
	return result, err
}

// YAML 格式保存数据

func SaveYaml(file string, data any) error {
	buf, err := yaml.Marshal(data)
	if err != nil {
		return err
	}
	return Write(file, buf)
}

func ReadYaml(file string, data any) error {
	bytes, err := Read(file)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(bytes, data)
}

func ReadYamlT[T any](file string) (T, error) {
	var result T
	bytes, err := Read(file)
	if err != nil {
		return result, err
	}
	err = yaml.Unmarshal(bytes, &result)
	return result, err
}
