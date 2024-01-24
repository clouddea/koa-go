package config

type Config struct {
	Server struct {
		Name string
		Port int
	}
	Database    string
	JsonMaxSize int `yaml:"jsonMaxSize"`
}
