package entity

type PathItem struct {
	FileName string `json:"fileName"`
	Path     string `json:"path"`
	IsFolder int    `json:"isFolder"`
	EditTime string `json:"editTime"`
	FileSize string `json:"fileSize"`
	FullPath string `json:"fullPath"`
	IsDanger int    `json:"isDanger"`
}
