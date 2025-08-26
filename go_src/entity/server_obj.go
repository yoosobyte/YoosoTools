package entity

type ServerObj struct {
	// 服务器Id
	ServerId int `json:"serverId"`
	// 服务器名称
	ServerName string `json:"serverName"`
	// 服务器地址
	ServerUrl string `json:"serverUrl"`
	// 服务器端口
	ServerPort string `json:"serverPort"`
	// 服务器账号
	ServerUserName string `json:"serverUserName"`
	// 服务器密码
	ServerPassword string `json:"serverPassword"`
}
