package config

import (
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

type Config struct {
	// mapstructure: chỉ dẫn dành riêng cho thư viện viper
	//cách hoạt động : nó bảo viper rằng khi đọc file config(vd: config.yaml) hãy tìm mục nào
	//có thên là server và đổ dữ liệu vào biến Server
	//  Tên biến  /kiểu dữ liệu
	Server    ServerConfig    `mapstructure:"server"`
	Database  DatabaseConfig  `mapstructure:"database"`
	Redis     RedisConfig     `mapstructure:"redis"`
	RabbitMQ  RabbitMQConfig  `mapstructure:"rabbitmq"`
	JWT       JWTConfig       `mapstructure:"jwt"`
	FlashSale FlashSaleConfig `mapstructure:"flash_sale"`
}

type ServerConfig struct {
	Port string `mapstructure:"port" validate:"required"`
	// - oneof: Chỉ chấp nhận 1 trong 3 giá trị: "debug", "release", hoặc "test"
	Mode string `mapstructure:"mode" validate:"oneof=debug release test" `
}

type DatabaseConfig struct {
	Host               string `mapstructure:"host" validate:"required"`
	Port               string `mapstructure:"port" validate:"required"`
	User               string `mapstructure:"user" validate:"required"`
	Password           string `mapstructure:"password" validate:"required"`
	DBName             string `mapstructure:"dbname" validate:"required"`
	SSLMode            string `mapstructure:"sslmode"`
	MaxConnections     int    `mapstructure:"max_connections"`
	MaxIdleConnections int    `mapstructure:"max_idle_connections"`
}

type RedisConfig struct {
	Host     string `mapstructure:"host" validate:"required"`
	Port     string `mapstructure:"port" validate:"required"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
	PoolSize int    `mapstructure:"pool_size"`
}

type RabbitMQConfig struct {
	URL           string `mapstructure:"url" validate:"required,url"`
	Exchange      string `mapstructure:"exchange" validate:"required"`
	Queue         string `mapstructure:"queue" validate:"required"`
	RoutingKey    string `mapstructure:"routing_key" validate:"required"`
	PrefetchCount int    `mapstructure:"prefetch_count"`
}
type JWTConfig struct {
	Secret      string `mapstructure:"secret" validate:"required,min=8"`
	ExpireHours int    `mapstructure:"expire_hours" validate:"required,min=1"`
}

type FlashSaleConfig struct {
	StockTTL     int `mapstructure:"stock_ttl"`
	OrderTimeout int `mapstructure:"order_timeout"`
}

func LoadConfig(path string) (*Config, error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	// 1. Setup Environment Variables
	// thay thế (.) thành (_) trong tên biến môi trường
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	//2. Đọc file cấu hình
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// file cấu hình không tồn tại
			return nil, err
		}
	}

	//3. unmarchal(đổ dữ liệu vào struct)
	//đây là bước chuyển đổi từ "dữ liệu thô" ủa Viper sang "Struct Go" có kiểu dữ liệu rõ ràng
	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	// 4 Validate (Kiểm tra chất lượng dữ liệu)
	validate := validator.New()
	if err := validate.Struct(&config); err != nil {
		return nil, err
	}
	return &config, nil
}
