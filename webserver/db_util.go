package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"

	_ "github.com/Go-SQL-Driver/MySQL"
)

type serverInfo struct {
	Host     string
	Cpus     string
	Username string
	Ower     string
	Endtime  string
}

func server_list() ([]byte, error) {
	mysql_handle, err := sql.Open("mysql", *mysql_dsn)
	if err != nil {
		log.Println("connect mysql err ", err)
		return nil, err
	}
	rows, err := mysql_handle.Query("select host_ip,cpus,username,ower,endtime from user")
	if err != nil {
		log.Println("get db result failed ", err)
		return nil, err
	} else {
		defer rows.Close()
		var res_json []serverInfo
		var host_, cpu_, username_, ower_, endtime_ string
		for rows.Next() {
			err := rows.Scan(&host_, &cpu_, &username_, &ower_, &endtime_)
			if err != nil {
				log.Println(err)
			}
			var tmp_info serverInfo
			tmp_info.Host = host_
			tmp_info.Cpus = cpu_
			tmp_info.Username = username_
			tmp_info.Ower = ower_
			tmp_info.Endtime = endtime_
			res_json = append(res_json, tmp_info)
		}
		s, _ := json.Marshal(res_json)
		log.Println("===>", string(s))
		return s, nil
	}
}
func applyServer(address string, username string, passwd string, endtime string) error {
	endtime = strings.Replace(endtime, "T", " ", 1)
	endtime = endtime + ":00"
	mysql_handle, err := sql.Open("mysql", *mysql_dsn)
	if err != nil {
		log.Println("connect mysql err ", err)
		return err
	}
	var ower string
	err = mysql_handle.QueryRow("select ower from user where host_ip = '" + address + "'").Scan(&ower)
	if err == nil && len(ower) > 0 {
		log.Println("user has exsist ", ower)
		return errors.New("user has exsist ")
	} else {
		log.Println("get owner ", ower)
		sql_str := fmt.Sprintf("update user set username='%s',ower='%s',passwd='%s',endtime='%s' where host_ip='%s'", username, username, passwd, endtime, address)
		log.Println(sql_str)
		if _, err = mysql_handle.Exec(sql_str); err != nil {
			return err
		}
		return nil
	}
}
