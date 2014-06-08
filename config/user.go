package config

import (
	"encoding/json"
	// "log"
	// "reflect"
)

type userAppConfig struct {
	source           string
	listen           string
	lruSize          uint64
	storageAppConfig StorageAppConfig
	indexAppConfig   IndexAppConfig
}

type userIndexAppConfig struct {
	engine        string
	localBasePath string
}

type userStorageAppConfig struct {
	engine    string
	basePath  string
	s3Key     string
	s3Secret  string
	s3Host    string
	s3Buckets []string
}

func NewUserAppConfig(content []byte) (AppConfig, error) {

	var f interface{}
	err := json.Unmarshal(content, &f)
	if err != nil {
		return nil, err
	}

	m := f.(map[string]interface{})
	appConfig := new(userAppConfig)
	appConfig.source = string(content)

	appConfig.listen, err = parseString("config", "listen", m)
	if err != nil {
		return nil, err
	}

	appConfig.lruSize, err = parseUint64("config", "lruSize", m)
	if err != nil {
		return nil, err
	}

	appConfig.indexAppConfig, err = newUserIndexAppConfig(m)
	if err != nil {
		return nil, err
	}

	appConfig.storageAppConfig, err = newUserStorageAppConfig(m)
	if err != nil {
		return nil, err
	}

	return appConfig, nil
}

func (c *userAppConfig) Source() string {
	return c.source
}

func (c *userAppConfig) Listen() string {
	return c.listen
}

func (c *userAppConfig) LruSize() uint64 {
	return c.lruSize
}

func (c *userAppConfig) Storage() StorageAppConfig {
	return c.storageAppConfig
}

func (c *userAppConfig) Index() IndexAppConfig {
	return c.indexAppConfig
}

func (c *userStorageAppConfig) Engine() string {
	return c.engine
}

func (c *userStorageAppConfig) BasePath() string {
	return c.basePath
}

func (c *userStorageAppConfig) S3Key() string {
	return c.s3Key
}

func (c *userStorageAppConfig) S3Secret() string {
	return c.s3Secret
}

func (c *userStorageAppConfig) S3Buckets() []string {
	return c.s3Buckets
}

func (c *userStorageAppConfig) S3Host() string {
	return c.s3Host
}

func (c *userIndexAppConfig) Engine() string {
	return c.engine
}

func (c *userIndexAppConfig) LocalBasePath() string {
	return c.localBasePath
}

func newUserStorageAppConfig(m map[string]interface{}) (StorageAppConfig, error) {
	data, err := parseConfigGroup("storage", m)
	if err != nil {
		return nil, err
	}

	config := new(userStorageAppConfig)

	config.engine, err = parseString("storage", "engine", data)
	if err != nil {
		return nil, err
	}

	if config.engine == "local" {
		config.basePath, err = parseString("storage", "basePath", data)
		if err != nil {
			return nil, err
		}
	}

	if config.engine == "s3" {
		config.s3Key, err = parseString("storage", "s3Key", data)
		if err != nil {
			return nil, err
		}
		config.s3Secret, err = parseString("storage", "s3Secret", data)
		if err != nil {
			return nil, err
		}
		config.s3Host, err = parseString("storage", "s3Host", data)
		if err != nil {
			return nil, err
		}
		config.s3Buckets, err = parseStringArray("storage", "s3Buckets", data)
		if err != nil {
			return nil, err
		}
	}

	return config, nil
}

func newUserIndexAppConfig(m map[string]interface{}) (IndexAppConfig, error) {
	data, err := parseConfigGroup("index", m)
	if err != nil {
		return nil, err
	}

	config := new(userIndexAppConfig)

	config.engine, err = parseString("index", "engine", data)
	if err != nil {
		return nil, err
	}

	if config.engine == "local" {
		config.localBasePath, err = parseString("index", "localBasePath", data)
		if err != nil {
			return nil, err
		}
	}

	return config, nil
}
