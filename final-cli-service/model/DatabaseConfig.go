package model

type DatabaseConfig struct {
	Name           string `json:"database_name"`
	Host           string `json:"db_host"`
	Port           string `json:"db_port"`
	Username       string `json:"db_username"`
	ID             int    `json:"id"`
	Error          error
	FileSQL        string 
	FileDownloaded string
}