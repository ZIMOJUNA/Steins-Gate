package config

// Config 总配置结构体（对应整个yaml文件）
type Config struct {
	Server  ServerConfig  `yaml:"server"`  // 映射yaml的server节点
	Redis   RedisConfig   `yaml:"redis"`   // 映射yaml的redis节点
	MongoDB MongoDBConfig `yaml:"mongodb"` // 映射yaml的mongodb节点
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port int    `yaml:"port"`
	Mode string `yaml:"mode"`
}

// RedisConfig Redis配置
type RedisConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

// MongoDBConfig MongoDB配置
type MongoDBConfig struct {
	Url    string `yaml:"url"`
	DBName string `yaml:"dbname"`
}
