package config

type ConnectionConfig struct {
	Username string   `json:"username"`
	Password string   `json:"password"`
	Port     int      `json:"port"`
	Adapter  string   `json:"adapter"`
	Host     string   `json:"host"`
	Database string   `json:"database"`
	Tables   []string `json:"tables"`
}
