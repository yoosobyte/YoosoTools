package entity

type LocalSet struct {
	BuildEducationCloudSet ProjectSet `json:"buildEducationCloudSet"`
	BuildEducationCloudLog ProjectLog `json:"buildEducationCloudLog"`
	KillPort               string     `json:"killPort"`
}

type ProjectLog struct {
	ModuleNameList       []string `json:"moduleNameList"`
	ServerPortStatus     string   `json:"serverPortStatus"`
	MavenSetStatus       string   `json:"mavenSetStatus"`
	PackageRate          float64  `json:"packageRate"`
	UploadRate           float64  `json:"uploadRate"`
	HookShellStatus      string   `json:"hookShellStatus"`
	ModuleNameStatusList []string `json:"moduleNameStatusList"`
}

type ProjectSet struct {
	// 项目名
	ProjectName string `json:"projectName"`
	// maven软件位置
	MavenSoftPath string `json:"mavenSoftPath"`
	// maven仓库位置
	MavenRepoPath string `json:"mavenRepoPath"`
	// java软件位置
	JavaSoftPath string `json:"javaSoftPath"`
	// 项目位置
	ProjectPath string `json:"projectPath"`
	// 主动修改为edu分支
	AutoEdu string `json:"autoEdu"`
	// 服务器地址
	ServerUrl string `json:"serverUrl"`
	// 服务器端口
	ServerPort string `json:"serverPort"`
	// 服务器账号
	ServerUserName string `json:"serverUserName"`
	// 服务器密码
	ServerPassword string `json:"serverPassword"`
	// hook-命令
	ShellHook string `json:"shellHook"`
}
