package config

type Credential struct {
	Region          string `yaml:"region"`
	AccessKeyID     string `yaml:"access_key_id"`
	AccessKeySecret string `yaml:"access_key_secret"`
	Endpoint        string `yaml:"endpoint"`
	Token           string `yaml:"token"`
}

