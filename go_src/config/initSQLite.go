package config

import (
	"YoosoTools/go_src/controller"
	"database/sql"
	"log"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite" // 匿名导入驱动
)

// DB 全局数据库连接实例
var DB *sql.DB

// InitDB 初始化数据库连接
func InitDB() error {
	// 创建应用专属目录
	err := os.MkdirAll(controller.YoosoToolsDir, 0755)
	if err != nil {
		return err
	}

	// 数据库文件路径
	dbPath := filepath.Join(controller.YoosoToolsDir, "data.db")
	log.Printf("SQLite数据库位置: %s", dbPath)

	// 打开（或创建）数据库
	database, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return err
	}

	// 测试连接
	if err = database.Ping(); err != nil {
		return err
	}

	// 设置连接池参数（可选）
	database.SetMaxOpenConns(1) // SQLite 写操作通常需要串行化，建议设为1
	DB = database

	// 创建服务器表
	if err = createTables(); err != nil {
		return err
	}

	log.Println("SQLite初始化成功")
	return nil
}

// createTables 创建数据表
func createTables() error {
	//dropTableSQL := `DROP TABLE IF EXISTS server;`
	//_, err1 := DB.Exec(dropTableSQL)
	//if err1 != nil {
	//	return fmt.Errorf("删除表失败: %v", err1)
	//}
	createTaskTable := `
    CREATE TABLE IF NOT EXISTS server (
        server_id INTEGER PRIMARY KEY AUTOINCREMENT,
        server_name TEXT NOT NULL,
        server_url TEXT NOT NULL,
        server_port TEXT NOT NULL,
        server_user_name TEXT NOT NULL,
        server_password TEXT NOT NULL
    );
    `
	_, err2 := DB.Exec(createTaskTable)
	return err2
}
