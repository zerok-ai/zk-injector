package config

type RedisConfig struct {
	Host        string `yaml:"host" env:"REDIS_HOST" env-description:"Database host"`
	Port        string `yaml:"port" env:"REDIS_PORT" env-description:"Database port"`
	ReadTimeout int    `yaml:"readTimeout"`
}

type WebhookConfig struct {
	Namespace string `yaml:"namespace"`
	Service   string `yaml:"service"`
	Name      string `yaml:"name"`
	Path      string `yaml:"path"`
}

type ZkInjectorConfig struct {
	Redis   RedisConfig   `yaml:"redis"`
	Webhook WebhookConfig `yaml:"webhook"`
	Debug   bool          `yaml:"debug"`
}
