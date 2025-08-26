package dao

import (
	"YoosoTools/go_src/config"
	"YoosoTools/go_src/entity"
	"database/sql"
	"log"
)

func SaveServerDb(serverObj entity.ServerObj) (entity.ServerObj, error) {
	result, err := config.DB.Exec(
		"INSERT INTO server (server_name,server_url, server_port, server_user_name, server_password) VALUES (?, ?, ?, ?, ?)",
		serverObj.ServerName, serverObj.ServerUrl, serverObj.ServerPort, serverObj.ServerUserName, serverObj.ServerPassword,
	)
	if err != nil {
		return entity.ServerObj{}, err // 返回空的ServerObj和错误
	}

	id, err := result.LastInsertId()
	if err != nil {
		return entity.ServerObj{}, err
	}

	serverObj.ServerId = int(id)
	return serverObj, nil
}

func GetListServerDb() ([]entity.ServerObj, error) {
	rows, err := config.DB.Query(`
        SELECT server_id, server_name, server_url, server_port, server_user_name, server_password
        FROM server 
        ORDER BY server_id DESC
    `)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var servers []entity.ServerObj
	for rows.Next() {
		var t entity.ServerObj
		err := rows.Scan(
			&t.ServerId, &t.ServerName, &t.ServerUrl, &t.ServerPort, &t.ServerUserName, &t.ServerPassword,
		)
		if err != nil {
			log.Println("Error scanning task:", err)
			continue
		}
		servers = append(servers, t)
	}

	return servers, nil
}

func GetOneServerDb(serverId int) (*entity.ServerObj, error) {
	var obj entity.ServerObj
	err := config.DB.QueryRow(`
        SELECT server_id,server_name, server_url, server_port, server_user_name, server_password
        FROM server 
        WHERE server_id = ?
    `, serverId).Scan(
		&obj.ServerId, &obj.ServerName, &obj.ServerUrl, &obj.ServerPort, &obj.ServerUserName, &obj.ServerPassword,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // 没有找到记录不是错误，返回 nil
		}
		return nil, err
	}
	return &obj, nil
}

func EditServerDb(serverObj entity.ServerObj) (entity.ServerObj, error) {
	result, err := config.DB.Exec(`
        UPDATE server 
        SET server_name = ?, server_url = ?, server_port = ?, server_user_name = ?, server_password = ? 
        WHERE server_id = ?
    `, serverObj.ServerName, serverObj.ServerUrl, serverObj.ServerPort, serverObj.ServerUserName, serverObj.ServerPassword, serverObj.ServerId)
	if err != nil {
		return entity.ServerObj{}, err // 返回空的ServerObj和错误
	}

	id, err := result.LastInsertId()
	if err != nil {
		return entity.ServerObj{}, err
	}

	serverObj.ServerId = int(id)
	return serverObj, nil
}

func RemoveServerDb(serverId int) error {
	_, err := config.DB.Exec("DELETE FROM server WHERE server_id = ?", serverId)
	return err
}
