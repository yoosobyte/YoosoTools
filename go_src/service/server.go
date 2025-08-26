package service

import (
	"YoosoTools/go_src/dao"
	"YoosoTools/go_src/entity"
	"encoding/json"
	"log"
)

func SaveServer(addStr string) string {
	var obj entity.ServerObj

	// 解析 JSON 字符串到结构体
	err := json.Unmarshal([]byte(addStr), &obj)
	if err != nil {
		log.Printf("JSON 解析失败: %v", err)
		return entity.ErrorOnlyMsgStr("JSON解析失败")
	}

	// 调用数据库保存方法
	savedObj, err := dao.SaveServerDb(obj)
	if err != nil {
		log.Printf("保存到数据库失败1: %v", err)
		return entity.ErrorOnlyMsgStr("数据库保存失败")
	}
	return entity.SuccessOnlyDataStr(savedObj)
}

func EditServer(editStr string) string {
	var obj entity.ServerObj
	// 解析 JSON 字符串到结构体
	err := json.Unmarshal([]byte(editStr), &obj)
	if err != nil {
		log.Printf("JSON 解析失败: %v", err)
		return entity.ErrorOnlyMsgStr("JSON解析失败")
	}

	// 调用数据库保存方法
	savedObj, err := dao.EditServerDb(obj)
	if err != nil {
		log.Printf("保存到数据库失败1: %v", err)
		return entity.ErrorOnlyMsgStr("数据库保存失败")
	}
	return entity.SuccessOnlyDataStr(savedObj)
}

func RemoveServer(serverId int) string {
	err := dao.RemoveServerDb(serverId)
	if err == nil {
		return entity.SuccessStr()
	} else {
		return entity.ErrorStr()
	}
}

func GetOneServer(serverId int) string {
	obj, err := dao.GetOneServerDb(serverId)
	if err == nil {
		return entity.SuccessOnlyDataStr(obj)
	} else {
		return entity.ErrorStr()
	}
}

func GetListServer() string {
	serverList, err := dao.GetListServerDb()
	if err != nil {
		log.Printf("保存到数据库失败2: %v", err)
		return entity.ErrorOnlyMsgStr("数据库保存失败")
	}
	return entity.SuccessOnlyDataStr(serverList)
}
