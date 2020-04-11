package common

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)
const timeLayout  = "20060102150405"
//定义接口接收数据 ps仅限龙虎
type LHPostData struct {
	UserId  int
	DeskId  int
	BootNum int
	PaveNum int
	Dragon  int
	Tie     int
	Tiger   int
}
//定义接口接收数据 ps仅限百家乐
type BJLPostData struct {
	UserId  int
	DeskId  int
	BootNum int
	PaveNum int
	Player  int
	PlayerPair int
	Tie   int
	BankerPair int
	Banker int
}
//定义返回数据格式
type result struct {
	Code int
	Msg  string
	Data map[string]interface{}
}
//返回数据json
func AjaxReturn(w http.ResponseWriter,code int,msg string ,data map[string]interface{}){
	arr:=&result{
		code,
		msg,
		data,
	}
	b, jsonErr := json.Marshal(arr) //json化结果集
	if jsonErr != nil {
		fmt.Fprintln(w,"encoding fail")
	}
	fmt.Fprintln(w,string(b))
	return
}
//无重复游戏订单号
func GetOrderSn(game string) string {
	t:=time.Now()
	pre:=t.Format(timeLayout)
	rand.Seed(time.Now().Unix())
	num:=rand.Intn(99999999-11111111+11111111) + 11111111 //[11111111,99999999]
	suf:=strconv.FormatInt(int64(num), 10)
	orderSn :=game+pre+suf
	return orderSn
}
//根据订单号获取表后缀
func GetTableSuf(orderSn string) string {
	suf := []rune(orderSn)
	return string(suf[1:9])
}
//根据当前时间获取表后缀
func TableSufByTime() string {
	t:=time.Now()
	pre:=t.Format(timeLayout)
	suf := []rune(pre)
	return string(suf[0:8])
}