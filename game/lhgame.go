package game

import (
	"../common"
	"../model"
	"database/sql"
	"net/http"
	"strconv"
)
//进桌
func InTable(w http.ResponseWriter, r *http.Request)  {
	w.Header().Set("Access-Control-Allow-Origin", "*") //允许跨域
	common.AjaxReturn(w,1,"进桌",nil)
	return
}
//离桌
func OutTable(w http.ResponseWriter, r *http.Request)  {
	w.Header().Set("Access-Control-Allow-Origin", "*") //允许跨域
	common.AjaxReturn(w,1,"离桌",nil)
	return
}
//确定下注
func LHSetBet(w http.ResponseWriter, r *http.Request)  {
	w.Header().Set("Access-Control-Allow-Origin", "*") //允许跨域
	var data common.LHPostData	//接受数据结构体 龙虎
	data.UserId,_= strconv.Atoi(r.Header.Get("user_id"))
	data.DeskId,_=strconv.Atoi(r.PostFormValue("desk_id"))
	data.BootNum,_=strconv.Atoi(r.PostFormValue("boot_num"))
	data.PaveNum,_=strconv.Atoi(r.PostFormValue("pave_num"))

	data.Dragon,_=strconv.Atoi(r.PostFormValue("dragon"))
	data.Tie,_=strconv.Atoi(r.PostFormValue("tie"))
	data.Tiger,_=strconv.Atoi(r.PostFormValue("tiger"))
	betMoney:=data.Dragon+data.Tie+data.Tiger

	var db *sql.DB = common.ConMysql()
	defer db.Close()

	balance:=model.GetUserBalance(db,data.UserId)/100
	if balance< betMoney {
		common.AjaxReturn(w,0,"余额不足",nil)
		return
	}

	limit:=model.GetLHDeskLimit(db,data.DeskId)
	if data.Dragon>0 {
		if data.Dragon<limit["minLimit"]||data.Dragon>limit["maxLimit"]{
			common.AjaxReturn(w,0,"下注失败",nil)
			return
		}else{
			data.Dragon=data.Dragon*100
		}
	}else{
		data.Dragon=0
	}
	if data.Tie>0 {
		if data.Tie<limit["tieMinLimit"]||data.Tie>limit["tieMaxLimit"]{
			common.AjaxReturn(w,0,"下注失败",nil)
			return
		}else{
			data.Tie=data.Tie*100
		}
	}else{
		data.Tie=0
	}
	if data.Tiger>0 {
		if data.Tiger<limit["minLimit"]||data.Tiger>limit["maxLimit"]{
			common.AjaxReturn(w,0,"下注失败",nil)
			return
		}else{
			data.Tiger=data.Tiger*100
		}
	}else{
		data.Tiger=0
	}
	code,msg:= model.LHCertainBet(db,data)
	common.AjaxReturn(w,code,msg,nil)
}
//取消下注
func LHQuitBet(w http.ResponseWriter, r *http.Request)  {
	w.Header().Set("Access-Control-Allow-Origin", "*") //允许跨域
	var data common.LHPostData
	data.UserId,_= strconv.Atoi(r.Header.Get("user_id"))
	data.DeskId,_=strconv.Atoi(r.PostFormValue("desk_id"))
	data.BootNum,_=strconv.Atoi(r.PostFormValue("boot_num"))
	data.PaveNum,_=strconv.Atoi(r.PostFormValue("pave_num"))
	var db *sql.DB = common.ConMysql()
	defer db.Close()
	code,msg:= model.LHCancelBet(db,data)
	common.AjaxReturn(w,code,msg,nil)
}
//确定游戏结果
func LHGameOver(w http.ResponseWriter, r *http.Request)  {
	w.Header().Set("Access-Control-Allow-Origin", "*") //允许跨域
	var data common.LHPostData
	data.DeskId,_=strconv.Atoi(r.PostFormValue("desk_id"))
	data.BootNum,_=strconv.Atoi(r.PostFormValue("boot_num"))
	data.PaveNum,_=strconv.Atoi(r.PostFormValue("pave_num"))
	gameNum,_:=strconv.Atoi(r.PostFormValue("game_num"))//7龙4虎1和
	go LHGameDone(data,gameNum)//游戏结算
	go LHRecord(data,gameNum)//游戏结果入库
}

//游戏结算
func LHGameDone(data common.LHPostData,gameNum int)  {
	var db *sql.DB = common.ConMysql()
	defer db.Close()
	model.GetLHUserIds(db,data,gameNum)
	return
}
//游戏记录入库
func LHRecord(data common.LHPostData,gameNum int) {
	var db *sql.DB = common.ConMysql()
	defer db.Close()
	model.LHgameRecord(db,data,gameNum)
	return
}