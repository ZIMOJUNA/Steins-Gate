package config

import "time"

// Config 总配置结构体（对应整个yaml文件）
type Config struct {
	Server ServerConfig `yaml:"server"` // 映射yaml的server节点
	Redis  RedisConfig  `yaml:"redis"`  // 映射yaml的redis节点
	MySQL  MySQLConfig  `yaml:"mysql"`  // 映射yaml的mysql节点
	Auth   AuthConfig   `yaml:"auth"`   // 映射yaml的auth节点
	Mail   MailConfig   `yaml:"mail"`   // 映射yaml的mail节点
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
	DSN      string `yaml:"dsn"`
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DBName   string `yaml:"dbname"`
}

// AuthConfig 账号、验证码和登录态配置。
type AuthConfig struct {
	TokenTTL                   string `yaml:"token_ttl"`
	EmailCodeTTL               string `yaml:"email_code_ttl"`
	EmailCodeResendInterval    string `yaml:"email_code_resend_interval"`
	EmailCodeSendLimitPerHour  int    `yaml:"email_code_send_limit_per_hour"`
	EmailCodeMaxVerifyAttempts int    `yaml:"email_code_max_verify_attempts"`
	EmailCodeHashSecret        string `yaml:"email_code_hash_secret"`
	PasswordMinLength          int    `yaml:"password_min_length"`
}

// MailConfig 邮件发送配置。provider 当前支持 console 和 smtp，aliyun 预留。
type MailConfig struct {
	Provider string     `yaml:"provider"`
	SMTP     SMTPConfig `yaml:"smtp"`
}

// SMTPConfig SMTP 邮件服务配置。
type SMTPConfig struct {
	Host       string `yaml:"host"`
	Port       int    `yaml:"port"`
	Username   string `yaml:"username"`
	Password   string `yaml:"password"`
	From       string `yaml:"from"`
	FromName   string `yaml:"from_name"`
	UseTLS     bool   `yaml:"use_tls"`
	StartTLS   bool   `yaml:"start_tls"`
	SkipVerify bool   `yaml:"skip_verify"`
}

func (c AuthConfig) TokenDuration() time.Duration {
	return parseDuration(c.TokenTTL, 24*time.Hour)
}

func (c AuthConfig) EmailCodeDuration() time.Duration {
	return parseDuration(c.EmailCodeTTL, 5*time.Minute)
}

func (c AuthConfig) ResendInterval() time.Duration {
	return parseDuration(c.EmailCodeResendInterval, time.Minute)
}

func (c AuthConfig) SendLimitPerHour() int {
	if c.EmailCodeSendLimitPerHour <= 0 {
		return 5
	}
	return c.EmailCodeSendLimitPerHour
}

func (c AuthConfig) MaxVerifyAttempts() int {
	if c.EmailCodeMaxVerifyAttempts <= 0 {
		return 5
	}
	return c.EmailCodeMaxVerifyAttempts
}

func (c AuthConfig) MinPasswordLength() int {
	if c.PasswordMinLength <= 0 {
		return 8
	}
	return c.PasswordMinLength
}

func parseDuration(value string, fallback time.Duration) time.Duration {
	if value == "" {
		return fallback
	}
	duration, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}
	return duration
}
