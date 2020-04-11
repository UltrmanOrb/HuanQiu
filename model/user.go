package model

import (
	"../common"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)
//时间戳
var nowtime int64=time.Now().Unix()
var code int
var msg string
//获取用户余额
func GetUserBalance(db *sql.DB,user_id int)int  {
	var balance int
	err:=db.QueryRow("select balance from hq_user_account where id=?", user_id).Scan(&balance)
	if err!=nil{
		fmt.Println(err)
	}
	return balance
}
//龙虎确定下注
func LHCertainBet(db *sql.DB, data common.LHPostData)(int,string){
	var deskName string
	deskName,_= getDeskName(db,data.DeskId)
	orderSn :=common.GetOrderSn("L")
	tableSuf:=common.GetTableSuf(orderSn)
	moneyMap :=make(map[string]int)
	moneyMap["dragon"]=data.Dragon
	moneyMap["tie"]=data.Tie
	moneyMap["tiger"]=data.Tiger
	bet,_:=json.Marshal(moneyMap)
	betMoney :=string(bet)
	allMoney :=data.Dragon+data.Tie+data.Tiger
	//开启事务
	tx,_:=db.Begin()
	defer tx.Rollback()
	//账变前
	var betBefore int64
	berr:=tx.QueryRow("select balance from hq_user_account where user_id=? for update ",data.UserId).Scan(&betBefore)
	if berr!=nil{
		fmt.Println(berr)
	}
	//余额减去下注金额
	account,err1:=tx.Exec("update hq_user_account set balance=balance-?,savetime=? where user_id=?", allMoney,nowtime,data.UserId)
	if err1!=nil{
		fmt.Println(err1)
	}
	num1,_:=account.RowsAffected()
	if num1==0{
		tx.Rollback()
		code=0
		msg="account fail"
		return code,msg
	}
	//账变后
	var betAfter int64
	aerr:=tx.QueryRow("select balance from hq_user_account where user_id=? for update ",data.UserId).Scan(&betAfter)
	if aerr!=nil{
		fmt.Println(aerr)
	}

	//插入用户流水记录
	billflow,err2:=tx.Exec("insert into hq_user_billflow_"+tableSuf+"(user_id,score,bet_before,bet_after,order_sn,status,game_type,remark,creatime) values (?,?,?,?,?,?,?,?,?)",data.UserId,-allMoney, betBefore, betAfter, orderSn,2,2,"龙虎下注冻结",nowtime)
	if err2!=nil{
		fmt.Println(err2)
	}
	num2,_:=billflow.RowsAffected()
	if num2==0{
		tx.Rollback()
		code=0
		msg="billflow fail"
		return code,msg
	}
	//插入游戏记录
	order,err3:=tx.Exec("insert into hq_order_"+tableSuf+"(user_id,order_sn,desk_id,desk_name,boot_num,pave_num,bet_money,status,creatime) values (?,?,?,?,?,?,?,?,?)",data.UserId, orderSn,data.DeskId, deskName,data.BootNum,data.PaveNum, betMoney,0,nowtime)
	if err3 != nil {
		fmt.Println(err3)
	}
	num3,_:=order.RowsAffected()
	if num3==0{
		tx.Rollback()
		code=0
		msg="order fail"
		return code,msg
	}
	tx.Commit()
	code=1
	msg="success"
	return code,msg
}
//龙虎取消下注
func LHCancelBet(db *sql.DB, data common.LHPostData) (int ,string) {
	tableSuf:=common.TableSufByTime()
	rows,err:=db.Query("select order_sn,bet_money from hq_order_"+tableSuf+" where user_id=? and desk_id=? and boot_num=? and pave_num=? and status=0",data.UserId,data.DeskId,data.BootNum,data.PaveNum)
	if err!=nil{
		fmt.Println(err)
	}
	moneyMap :=make(map[string]int)
	//开启事务
	tx,_:=db.Begin()
	defer tx.Rollback()
	for rows.Next(){
		var orderSn string
		var money string
		err= rows.Scan(&orderSn,&money)
		if err != nil {
			fmt.Println(err)
		}
		json.Unmarshal([]byte(money), &moneyMap)
		allMoney := moneyMap["dragon"]+ moneyMap["tie"]+ moneyMap["tiger"]

		//账变前
		var betBefore int64
		berr:=tx.QueryRow("select balance from hq_user_account where user_id=? for update ",data.UserId).Scan(&betBefore)
		if berr!=nil{
			fmt.Println(berr)
		}
		//账变
		account,err1:=tx.Exec("update hq_user_account set balance=balance+?,savetime=? where user_id=?", allMoney,nowtime,data.UserId)
		if err1!=nil{
			fmt.Println(err1)
		}
		num1,_:=account.RowsAffected()
		if num1==0{
			tx.Rollback()
			code=0
			msg="account fail"
			return code,msg
		}
		//账变后
		var betAfter int64
		aerr:=tx.QueryRow("select balance from hq_user_account where user_id=? for update ",data.UserId).Scan(&betAfter)
		if aerr!=nil{
			fmt.Println(aerr)
		}
		//插入用户流水记录
		billflow,err2:=tx.Exec("insert into hq_user_billflow_"+tableSuf+"(user_id,score,bet_before,bet_after,order_sn,status,game_type,remark,creatime) values (?,?,?,?,?,?,?,?,?)",data.UserId, allMoney, betBefore, betAfter, orderSn,2,2,"龙虎取消下注",nowtime)
		if err2!=nil{
			fmt.Println(err2)
		}
		num2,_:=billflow.RowsAffected()
		if num2==0{
			tx.Rollback()
			code=0
			msg="billflow fail"
			return code,msg
		}
		//更改游戏状态
		order,err3:=tx.Exec("update hq_order_"+tableSuf+" set status=2 where user_id=? and order_sn=?",data.UserId, orderSn)
		if err3!=nil{
			fmt.Println(err3)
		}
		num3,_:=order.RowsAffected()
		if num3==0{
			tx.Rollback()
			code=0
			msg="order fail"
			return code,msg
		}

	}
	tx.Commit()
	code=1
	msg="success"
	return code,msg
}
//遍历user_id 推送龙虎游戏结果
func GetLHUserIds(db *sql.DB, data common.LHPostData,gameNum int)  {
	tablesuf:=common.TableSufByTime()
	rows,err:=db.Query("select distinct(user_id) from hq_order_"+tablesuf+" where desk_id=? and boot_num=? and pave_num=? and status=0 group by user_id",data.DeskId,data.BootNum,data.PaveNum)
	if err!=nil{
		fmt.Println(err)
	}
	var sumMoney int
	var userId int
	for rows.Next(){
		err=rows.Scan(&userId)
		if err!=nil{
			fmt.Println(err)
		}
		code,msg,sumMoney=lhGameDo(db,data,gameNum,userId)
		fmt.Println(code,msg)
		fmt.Println("sendTo-->",userId)
		fmt.Println("sumMoney:",sumMoney)
		fmt.Println("------------")
	}
	return
}
//龙虎游戏结算
func lhGameDo(db *sql.DB, data common.LHPostData, gameNum int,userId int) (int,string,int) {
	tablesuf:=common.TableSufByTime()
	rows,err:=db.Query("select order_sn,bet_money from hq_order_"+tablesuf+" where user_id=? and desk_id=? and boot_num=? and pave_num=? and status=0",userId,data.DeskId,data.BootNum,data.PaveNum)
	if err!=nil{
		fmt.Println(err)
	}
	moneyMap :=make(map[string]int)
	sumMoney:=0
	tx,_:=db.Begin()
	defer tx.Rollback()
	for rows.Next(){
		var money ,orderSn string
		err= rows.Scan(&orderSn,&money)
		if err != nil {
			fmt.Println(err)
		}
		json.Unmarshal([]byte(money), &moneyMap)
		dragonMoney:=win(gameNum,7,moneyMap["dragon"])
		tieMoney:=win(gameNum,1,moneyMap["tie"])
		tigerMoney:=win(gameNum,4,moneyMap["tiger"])

		betMoney:=moneyMap["dragon"]+moneyMap["tie"]+moneyMap["tiger"] //下注总金额
		gameMoney:=dragonMoney+tieMoney+tigerMoney	//输赢总金额
		sumMoney=sumMoney+gameMoney	//推送总金额

		fmt.Println("下注金额",betMoney)
		fmt.Println("输赢金额",gameMoney)
		fmt.Println("推送金额",sumMoney)
		//解冻前
		var freeBefore int64
		fberr:=tx.QueryRow("select balance from hq_user_account where user_id=? for update ",userId).Scan(&freeBefore)
		if fberr!=nil{
			fmt.Println(fberr)
		}
		//解冻下注金额
		unfree,unerr:=tx.Exec("update hq_user_account set balance=balance+?,savetime=? where user_id=?", betMoney,nowtime,userId)
		if unerr!=nil{
			fmt.Println(unerr)
		}
		fnum,_:=unfree.RowsAffected()
		if fnum==0{
			tx.Rollback()
			code=0
			msg="unfree fail"
			return code,msg,0
		}
		//解冻后
		var freeAfter int64
		faerr:=tx.QueryRow("select balance from hq_user_account where user_id=? for update ",userId).Scan(&freeAfter)
		if faerr!=nil{
			fmt.Println(faerr)
		}
		//插入解冻流水
		unfreeBill,unerr:=tx.Exec("insert into hq_user_billflow_"+tablesuf+"(user_id,score,bet_before,bet_after,order_sn,status,game_type,remark,creatime) values (?,?,?,?,?,?,?,?,?)",userId, betMoney, freeBefore, freeAfter, orderSn,2,2,"龙虎下注解冻",nowtime)
		if unerr!=nil{
			fmt.Println(unerr)
		}
		unfreeNum,_:=unfreeBill.RowsAffected()
		if unfreeNum==0{
			tx.Rollback()
			code=0
			msg="unfreeBill fail"
			return code,msg,0
		}
		//开始结算
		//账变前
		var betBefore int64
		berr:=tx.QueryRow("select balance from hq_user_account where user_id=? for update ",userId).Scan(&betBefore)
		if berr!=nil{
			fmt.Println(berr)
		}

		//如果为和且未下和
		if gameMoney==0{
			//插入用户流水记录
			tiebill,terr1:=tx.Exec("insert into hq_user_billflow_"+tablesuf+"(user_id,score,bet_before,bet_after,order_sn,status,game_type,remark,creatime) values (?,?,?,?,?,?,?,?,?)",userId, gameMoney, betBefore, betBefore, orderSn,2,2,"龙虎游戏结算",nowtime)
			if terr1!=nil{
				fmt.Println(terr1)
			}
			tnum1,_:=tiebill.RowsAffected()
			if tnum1==0{
				tx.Rollback()
				code=0
				msg="tbillflow fail"
				return code,msg,0
			}
			//更改游戏状态
			torder,terr2:=tx.Exec("update hq_order_"+tablesuf+" set status=1 ,get_money=? where user_id=? and order_sn=?",sumMoney,userId, orderSn)
			if terr2!=nil{
				fmt.Println(terr2)
			}
			tnum2,_:=torder.RowsAffected()
			if tnum2==0{
				tx.Rollback()
				code=0
				msg="torder fail"
				return code,msg,0
			}
			tx.Commit()
			code=1
			msg="success"
			return code,msg,sumMoney
		}

		//账变
		account,err1:=tx.Exec("update hq_user_account set balance=balance+?,savetime=? where user_id=?", gameMoney,nowtime,userId)
		if err1!=nil{
			fmt.Println(err1)
		}
		num1,_:=account.RowsAffected()
		if num1==0{
			tx.Rollback()
			code=0
			msg="account fail"
			return code,msg,0
		}

		//账变后
		var betAfter int64
		aerr:=tx.QueryRow("select balance from hq_user_account where user_id=? for update ",userId).Scan(&betAfter)
		if aerr!=nil{
			fmt.Println(aerr)
		}
		//插入用户流水记录
		billflow,err2:=tx.Exec("insert into hq_user_billflow_"+tablesuf+"(user_id,score,bet_before,bet_after,order_sn,status,game_type,remark,creatime) values (?,?,?,?,?,?,?,?,?)",userId, gameMoney, betBefore, betAfter, orderSn,2,2,"龙虎游戏结算",nowtime)
		if err2!=nil{
			fmt.Println(err2)
		}
		num2,_:=billflow.RowsAffected()
		if num2==0{
			tx.Rollback()
			code=0
			msg="billflow fail"
			return code,msg,0
		}
		//更改游戏状态
		order,err3:=tx.Exec("update hq_order_"+tablesuf+" set status=1 ,get_money=? where user_id=? and order_sn=?",sumMoney,userId, orderSn)
		if err3!=nil{
			fmt.Println(err3)
		}
		num3,_:=order.RowsAffected()
		if num3==0{
			tx.Rollback()
			code=0
			msg="order fail"
			return code,msg,0
		}
	}
	tx.Commit()
	code=1
	msg="success"
	return code,msg,sumMoney
}
//龙虎游戏结果入库
func LHgameRecord(db *sql.DB, data common.LHPostData,gameNum int)error  {
	tableSuf:=common.TableSufByTime()
	_,err:=db.Exec("insert into hq_lhgame_record_"+tableSuf+"(desk_id,boot_num,pave_num,status,winner,creatime,endtime) values(?,?,?,?,?,?,?)",data.DeskId,data.BootNum,data.PaveNum,1,gameNum,nowtime,nowtime)
	if err!=nil{
		fmt.Println(err)
		return err
	}
	return nil
}
//百家乐确定下注
func BjlCertainBet(db *sql.DB, data common.BJLPostData)(int,string){
	var deskName string
	deskName,_= getDeskName(db,data.DeskId)
	orderSn :=common.GetOrderSn("B")
	tableSuf:=common.GetTableSuf(orderSn)
	moneyMap :=make(map[string]int)

	moneyMap["player"]=data.Player
	moneyMap["playerPair"]=data.PlayerPair
	moneyMap["tie"]=data.Tie
	moneyMap["banker"]=data.Banker
	moneyMap["bankerPair"]=data.BankerPair
	bet,_:=json.Marshal(moneyMap)
	betMoney :=string(bet)
	allMoney :=data.Player+data.PlayerPair+data.Tie+data.Banker+data.BankerPair

	//开启事务
	tx,_:=db.Begin()
	defer tx.Rollback()
	//账变前
	var betBefore int64
	berr:=tx.QueryRow("select balance from hq_user_account where user_id=? for update ",data.UserId).Scan(&betBefore)
	if berr!=nil{
		fmt.Println(berr)
	}
	//余额-下注金额
	account,err1:=tx.Exec("update hq_user_account set balance=balance-?,savetime=? where user_id=?", allMoney,nowtime,data.UserId)
	if err1!=nil{
		fmt.Println(err1)
	}
	num1,_:=account.RowsAffected()
	if num1==0{
		tx.Rollback()
		code=0
		msg="account fail"
		return code,msg
	}
	//账变后
	var betAfter int64
	aerr:=tx.QueryRow("select balance from hq_user_account where user_id=? for update ",data.UserId).Scan(&betAfter)
	if aerr!=nil{
		fmt.Println(aerr)
	}

	//插入用户流水记录
	billflow,err2:=tx.Exec("insert into hq_user_billflow_"+tableSuf+"(user_id,score,bet_before,bet_after,order_sn,status,game_type,remark,creatime) values (?,?,?,?,?,?,?,?,?)",data.UserId,-allMoney, betBefore, betAfter, orderSn,2,1,"百家乐下注冻结",nowtime)
	if err2!=nil{
		fmt.Println(err2)
	}
	num2,_:=billflow.RowsAffected()
	if num2==0{
		tx.Rollback()
		code=0
		msg="billflow fail"
		return code,msg
	}
	//插入游戏记录
	order,err3:=tx.Exec("insert into hq_order_"+tableSuf+"(user_id,order_sn,desk_id,desk_name,boot_num,pave_num,bet_money,status,creatime) values (?,?,?,?,?,?,?,?,?)",data.UserId, orderSn,data.DeskId, deskName,data.BootNum,data.PaveNum, betMoney,0,nowtime)
	if err3 != nil {
		fmt.Println(err3)
	}
	num3,_:=order.RowsAffected()
	if num3==0{
		tx.Rollback()
		code=0
		msg="order fail"
		return code,msg
	}
	tx.Commit()
	code=1
	msg="success"
	return code,msg
}
//百家乐取消下注
func BjlCancelBet(db *sql.DB, data common.BJLPostData) (int ,string) {
	tableSuf:=common.TableSufByTime()
	rows,err:=db.Query("select order_sn,bet_money from hq_order_"+tableSuf+" where user_id=? and desk_id=? and boot_num=? and pave_num=? and status=0",data.UserId,data.DeskId,data.BootNum,data.PaveNum)
	if err!=nil{
		fmt.Println(err)
	}
	moneyMap :=make(map[string]int)
	//开启事务
	tx,_:=db.Begin()
	defer tx.Rollback()
	for rows.Next(){
		var orderSn string
		var money string
		err= rows.Scan(&orderSn,&money)
		if err != nil {
			fmt.Println(err)
		}
		json.Unmarshal([]byte(money), &moneyMap)
		allMoney := moneyMap["player"]+ moneyMap["playerPair"]+moneyMap["tie"]+ moneyMap["bankerPair"]+moneyMap["banker"]

		//账变前
		var betBefore int64
		berr:=tx.QueryRow("select balance from hq_user_account where user_id=? for update ",data.UserId).Scan(&betBefore)
		if berr!=nil{
			fmt.Println(berr)
		}
		//账变
		account,err1:=tx.Exec("update hq_user_account set balance=balance+?,savetime=? where user_id=?", allMoney,nowtime,data.UserId)
		if err1!=nil{
			fmt.Println(err1)
		}
		num1,_:=account.RowsAffected()
		if num1==0{
			tx.Rollback()
			code=0
			msg="account fail"
			return code,msg
		}
		//账变后
		var betAfter int64
		aerr:=tx.QueryRow("select balance from hq_user_account where user_id=? for update ",data.UserId).Scan(&betAfter)
		if aerr!=nil{
			fmt.Println(aerr)
		}
		//插入用户流水记录
		billflow,err2:=tx.Exec("insert into hq_user_billflow_"+tableSuf+"(user_id,score,bet_before,bet_after,order_sn,status,game_type,remark,creatime) values (?,?,?,?,?,?,?,?,?)",data.UserId, allMoney, betBefore, betAfter, orderSn,2,1,"百家乐取消下注",nowtime)
		if err2!=nil{
			fmt.Println(err2)
		}
		num2,_:=billflow.RowsAffected()
		if num2==0{
			tx.Rollback()
			code=0
			msg="billflow fail"
			return code,msg
		}
		//更改游戏状态
		order,err3:=tx.Exec("update hq_order_"+tableSuf+" set status=2 where user_id=? and order_sn=?",data.UserId, orderSn)
		if err3!=nil{
			fmt.Println(err3)
		}
		num3,_:=order.RowsAffected()
		if num3==0{
			tx.Rollback()
			code=0
			msg="order fail"
			return code,msg
		}

	}
	tx.Commit()
	code=1
	msg="success"
	return code,msg
}
//遍历user_id 推送百家乐游戏结果
func GetBjlUserIds(db *sql.DB, data common.BJLPostData,gameNum string)  {
	tablesuf:=common.TableSufByTime()
	rows,err:=db.Query("select distinct(user_id) from hq_order_"+tablesuf+" where desk_id=? and boot_num=? and pave_num=? and status=0 group by user_id",data.DeskId,data.BootNum,data.PaveNum)
	if err!=nil{
		fmt.Println(err)
	}
	var sumMoney int
	var userId int
	for rows.Next(){
		err=rows.Scan(&userId)
		if err!=nil{
			fmt.Println(err)
		}
		code,msg,sumMoney=bjlGameDo(db,data,gameNum,userId)
		fmt.Println(code,msg)
		fmt.Println("sendTo-->",userId)
		fmt.Println("sumMoney:",sumMoney)
		fmt.Println("------------")
	}
	return
}
//百家乐游戏结算
func bjlGameDo(db *sql.DB, data common.BJLPostData, gameNum string,userId int) (int,string,int) {
	tablesuf:=common.TableSufByTime()
	rows,err:=db.Query("select order_sn,bet_money from hq_order_"+tablesuf+" where user_id=? and desk_id=? and boot_num=? and pave_num=? and status=0",userId,data.DeskId,data.BootNum,data.PaveNum)
	if err!=nil{
		fmt.Println(err)
	}
	keyboardMap :=make(map[string]int)
	moneyMap :=make(map[string]int)
	sumMoney:=0
	var playerMoney,playerPairMoney,tieMoney,bankerPairMoney,bankerMoney int
	tx,_:=db.Begin()
	defer tx.Rollback()
	for rows.Next(){
		var money ,orderSn string
		err= rows.Scan(&orderSn,&money)
		if err != nil {
			fmt.Println(err)
		}
		json.Unmarshal([]byte(string(gameNum)), &keyboardMap)
		json.Unmarshal([]byte(money), &moneyMap)

		if len(keyboardMap)==1{
			fmt.Println("无对子")
			playerMoney=bjlGameCount(keyboardMap["game"],4,moneyMap["player"])
			tieMoney=bjlGameCount(keyboardMap["game"],1,moneyMap["tie"])
			bankerMoney=bjlGameCount(keyboardMap["game"],7,moneyMap["banker"])
			playerPairMoney=-moneyMap["playerPair"]
			bankerPairMoney=-moneyMap["bankerPair"]
		}else if len(keyboardMap)==2{
			fmt.Println("一对子")
			playerMoney=bjlGameCount(keyboardMap["game"],4,moneyMap["player"])
			tieMoney=bjlGameCount(keyboardMap["game"],1,moneyMap["tie"])
			bankerMoney=bjlGameCount(keyboardMap["game"],7,moneyMap["banker"])
			if _,ok:=keyboardMap["playerPair"];ok{
				playerPairMoney=moneyMap["playerPair"]*110/100
				bankerPairMoney=-moneyMap["bankerPair"]
			}else if _,ok:=keyboardMap["bankerPair"];ok{
				playerPairMoney=-moneyMap["playerPair"]
				bankerPairMoney=moneyMap["bankerPair"]*110/100
			}

		}else if len(keyboardMap)==3{
			fmt.Println("两对子")
			playerMoney=bjlGameCount(keyboardMap["game"],4,moneyMap["player"])
			tieMoney=bjlGameCount(keyboardMap["game"],1,moneyMap["tie"])
			bankerMoney=bjlGameCount(keyboardMap["game"],7,moneyMap["banker"])
			playerPairMoney=moneyMap["playerPair"]*110/100
			bankerPairMoney=moneyMap["bankerPair"]*110/100
		}
		betMoney:=moneyMap["player"]+moneyMap["playerPair"]+moneyMap["tie"]+moneyMap["bankerPair"]+moneyMap["banker"]//下注总金额
		gameMoney:=playerMoney+playerPairMoney+tieMoney+bankerPairMoney+bankerMoney	//输赢总金额
		sumMoney=sumMoney+gameMoney	//推送总金额

		fmt.Println("下注金额",betMoney)
		fmt.Println("输赢金额",gameMoney)
		fmt.Println("推送金额",sumMoney)
		//解冻前
		var freeBefore int64
		fberr:=tx.QueryRow("select balance from hq_user_account where user_id=? for update ",userId).Scan(&freeBefore)
		if fberr!=nil{
			fmt.Println(fberr)
		}
		//解冻下注金额
		unfree,unerr:=tx.Exec("update hq_user_account set balance=balance+?,savetime=? where user_id=?", betMoney,nowtime,userId)
		if unerr!=nil{
			fmt.Println(unerr)
		}
		fnum,_:=unfree.RowsAffected()
		if fnum==0{
			tx.Rollback()
			code=0
			msg="unfree fail"
			return code,msg,0
		}
		//解冻后
		var freeAfter int64
		faerr:=tx.QueryRow("select balance from hq_user_account where user_id=? for update ",userId).Scan(&freeAfter)
		if faerr!=nil{
			fmt.Println(faerr)
		}
		//插入解冻流水
		unfreeBill,unerr:=tx.Exec("insert into hq_user_billflow_"+tablesuf+"(user_id,score,bet_before,bet_after,order_sn,status,game_type,remark,creatime) values (?,?,?,?,?,?,?,?,?)",userId, betMoney, freeBefore, freeAfter, orderSn,2,2,"龙虎下注解冻",nowtime)
		if unerr!=nil{
			fmt.Println(unerr)
		}
		unfreeNum,_:=unfreeBill.RowsAffected()
		if unfreeNum==0{
			tx.Rollback()
			code=0
			msg="unfreeBill fail"
			return code,msg,0
		}
		//开始结算
		//账变前
		var betBefore int64
		berr:=tx.QueryRow("select balance from hq_user_account where user_id=? for update ",userId).Scan(&betBefore)
		if berr!=nil{
			fmt.Println(berr)
		}

		//如果为和且未下和
		if gameMoney==0{
			//插入用户流水记录
			tiebill,terr1:=tx.Exec("insert into hq_user_billflow_"+tablesuf+"(user_id,score,bet_before,bet_after,order_sn,status,game_type,remark,creatime) values (?,?,?,?,?,?,?,?,?)",userId, gameMoney, betBefore, betBefore, orderSn,2,2,"龙虎游戏结算",nowtime)
			if terr1!=nil{
				fmt.Println(terr1)
			}
			tnum1,_:=tiebill.RowsAffected()
			if tnum1==0{
				tx.Rollback()
				code=0
				msg="tbillflow fail"
				return code,msg,0
			}
			//更改游戏状态
			torder,terr2:=tx.Exec("update hq_order_"+tablesuf+" set status=1,get_money=? where user_id=? and order_sn=?",sumMoney,userId, orderSn)
			if terr2!=nil{
				fmt.Println(terr2)
			}
			tnum2,_:=torder.RowsAffected()
			if tnum2==0{
				tx.Rollback()
				code=0
				msg="torder fail"
				return code,msg,0
			}
			tx.Commit()
			code=1
			msg="success"
			return code,msg,sumMoney
		}

		//账变
		account,err1:=tx.Exec("update hq_user_account set balance=balance+?,savetime=? where user_id=?", gameMoney,nowtime,userId)
		if err1!=nil{
			fmt.Println(err1)
		}
		num1,_:=account.RowsAffected()
		if num1==0{
			tx.Rollback()
			code=0
			msg="account fail"
			return code,msg,0
		}

		//账变后
		var betAfter int64
		aerr:=tx.QueryRow("select balance from hq_user_account where user_id=? for update ",userId).Scan(&betAfter)
		if aerr!=nil{
			fmt.Println(aerr)
		}
		//插入用户流水记录
		billflow,err2:=tx.Exec("insert into hq_user_billflow_"+tablesuf+"(user_id,score,bet_before,bet_after,order_sn,status,game_type,remark,creatime) values (?,?,?,?,?,?,?,?,?)",userId, gameMoney, betBefore, betAfter, orderSn,2,2,"龙虎游戏结算",nowtime)
		if err2!=nil{
			fmt.Println(err2)
		}
		num2,_:=billflow.RowsAffected()
		if num2==0{
			tx.Rollback()
			code=0
			msg="billflow fail"
			return code,msg,0
		}
		//更改游戏状态
		order,err3:=tx.Exec("update hq_order_"+tablesuf+" set status=1 ,get_money=? where user_id=? and order_sn=?",sumMoney,userId, orderSn)
		if err3!=nil{
			fmt.Println(err3)
		}
		num3,_:=order.RowsAffected()
		if num3==0{
			tx.Rollback()
			code=0
			msg="order fail"
			return code,msg,0
		}
	}
	tx.Commit()
	code=1
	msg="success"
	return code,msg,sumMoney
}


//百家乐游戏结果入库
func BJLgameRecord(db *sql.DB, data common.BJLPostData,gameNum string)error  {
	tableSuf:=common.TableSufByTime()
	_,err:=db.Exec("insert into hq_bjlgame_record_"+tableSuf+"(desk_id,boot_num,pave_num,status,winner,creatime,endtime) values(?,?,?,?,?,?,?)",data.DeskId,data.BootNum,data.PaveNum,1,gameNum,nowtime,nowtime)
	if err!=nil{
		fmt.Println(err)
		return err
	}
	return nil
}
//获取龙虎台桌限红
func GetLHDeskLimit(db *sql.DB, deskId int)map[string]int  {
	limitMap:=make(map[string]int)
	var minLimit,minTieLimit,maxLimit,maxTieLimit string
	err:=db.QueryRow("select min_limit,max_limit,min_tie_limit,max_tie_limit from hq_desk where id=?", deskId).Scan(&minLimit,&maxLimit,&minTieLimit,&maxTieLimit)
	if err!=nil{
		fmt.Println(err)
	}
	minMap :=make(map[string]int)
	json.Unmarshal([]byte(minLimit), &minMap)
	maxMap :=make(map[string]int)
	json.Unmarshal([]byte(maxLimit), &maxMap)
	tieMinMap :=make(map[string]int)
	json.Unmarshal([]byte(minTieLimit), &tieMinMap)
	tieMaxMap :=make(map[string]int)
	json.Unmarshal([]byte(maxTieLimit), &tieMaxMap)
	limitMap["minLimit"]=minMap["c"]/100
	limitMap["maxLimit"]=maxMap["c"]/100
	limitMap["tieMinLimit"]=tieMinMap["c"]/100
	limitMap["tieMaxLimit"]=tieMaxMap["c"]/100
	return limitMap
}
//获取百家乐台桌限红
func GetBjlDeskLimit(db *sql.DB, deskId int)map[string]int  {
	limitMap:=make(map[string]int)
	var minLimit,minTieLimit,maxLimit,maxTieLimit ,minPairLimit,maxPairLimit string
	err:=db.QueryRow("select min_limit,max_limit,min_tie_limit,max_tie_limit ,min_pair_limit,max_pair_limit from hq_desk where id=?", deskId).Scan(&minLimit,&maxLimit,&minTieLimit,&maxTieLimit,&minPairLimit,&maxPairLimit)
	if err!=nil{
		fmt.Println(err)
	}
	minMap :=make(map[string]int)
	json.Unmarshal([]byte(minLimit), &minMap)
	maxMap :=make(map[string]int)
	json.Unmarshal([]byte(maxLimit), &maxMap)
	tieMinMap :=make(map[string]int)
	json.Unmarshal([]byte(minTieLimit), &tieMinMap)
	tieMaxMap :=make(map[string]int)
	json.Unmarshal([]byte(maxTieLimit), &tieMaxMap)
	pairMinMap :=make(map[string]int)
	json.Unmarshal([]byte(minPairLimit), &pairMinMap)
	pairMaxMap :=make(map[string]int)
	json.Unmarshal([]byte(maxPairLimit), &pairMaxMap)
	limitMap["minLimit"]=minMap["c"]/100
	limitMap["maxLimit"]=maxMap["c"]/100
	limitMap["tieMinLimit"]=tieMinMap["c"]/100
	limitMap["tieMaxLimit"]=tieMaxMap["c"]/100
	limitMap["pairMinLimit"]=pairMinMap["c"]/100
	limitMap["pairMaxLimit"]=pairMaxMap["c"]/100
	return limitMap
}
//获取龙虎台桌名字
func getDeskName(db *sql.DB, deskId int) (string,error) {
	desk:="hq_desk"
	var deskName string
	err:=db.QueryRow("select desk_name from "+desk+" where id=?", deskId).Scan(&deskName)
	if err!=nil{
		fmt.Println(err)
		return "",err
	}
	return deskName,err
}
//龙虎游戏输赢
func win(keybord int,betNum,money int)  int{
	var wallet int
	switch  {
	case keybord==7:
		if betNum==keybord{
			wallet=money*97/100
		}else{
			wallet=-money
		}
	case keybord==4:
		if betNum==keybord{
			wallet=money*97/100
		}else{
			wallet=-money
		}
	case keybord==1:
		if betNum==keybord{
			wallet=money*800/100
		}else{
			wallet=0
		}
	}
	return wallet
}
//百家乐游戏结算(无对子)
func bjlGameCount(key ,coin int,money int)int  {
	var wallet int
	switch  {
	case key==4:
		if key==coin{
			wallet=money
		}else{
			wallet=-money
		}
	case key==1:
		if key==coin{
			wallet=money*800/100
		}else{
			wallet=0
		}
	case key==7:
		if key==coin{
			wallet=money*95/100
		}else{
			wallet=-money
		}
	}
	return wallet
}