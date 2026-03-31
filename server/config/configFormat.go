package config

// Config 总配置结构体（对应整个yaml文件）
type Config struct {
	Server ServerConfig `yaml:"server"` // 映射yaml的server节点
	Redis  RedisConfig  `yaml:"redis"`  // 映射yaml的redis节点
	MySQL  MySQLConfig  `yaml:"mysql"`  // 映射yaml的mysql节点
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

// MySQLConfig MySQL配置
type MySQLConfig struct {
	DSN string `yaml:"dsn"`
}
