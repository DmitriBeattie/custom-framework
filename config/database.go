package config

type DatabaseConfig struct {
	Clusters           []Cluster `json:"clusters"`
	MaxOpenConnections int       `json:"maxOpenConn"`
	MaxIdleConnections int       `json:"maxIdleConnections"`
	User               string    `json:"user"`
	Password           string    `json:"pass"`
	DatabaseName       string    `json:"databaseName"`
}

type Cluster struct {
	Url
	IsReadonly        bool `json:"isReadonly"`
	IsFailoverPartner bool `json:"isFailoverPartner"`
}

func (d *DatabaseConfig) InstanceKind() string {
	return "database"
}
