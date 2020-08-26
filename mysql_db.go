package main

import (
	"database/sql"
	"fmt"
	"log"
	"runtime"

	_ "github.com/Go-SQL-Driver/MySQL"
)

func get_time() (string, error) {
	mysql_handle, err := sql.Open("mysql", *mysql_dsn)
	if err != nil {
		log.Println("connect mysql err ", err)
		return "", err
	}
	var time string
	err = mysql_handle.QueryRow("SELECT now()").Scan(&time)
	if err != nil {
		log.Println("get db result failed ", err)
		return "", err
	} else {
		return time, nil
	}
}
func get_endtime(ipaddr string) ([]byte, error) {
	mysql_handle, err := sql.Open("mysql", *mysql_dsn)
	if err != nil {
		log.Println("connect mysql err ", err)
		return nil, err
	}
	var endtime string
	err = mysql_handle.QueryRow("SELECT endtime from user where host_ip = '" + ipaddr + "'").Scan(&endtime)
	if err != nil {
		log.Println("get db result failed ", err)
		return nil, err
	} else {
		return []byte(endtime), nil
	}
}
func get_pass(user string) ([]byte, error) {
	mysql_handle, err := sql.Open("mysql", *mysql_dsn)
	if err != nil {
		log.Println("connect mysql err ", err)
		return nil, err
	}
	var username, passwd string
	err = mysql_handle.QueryRow("SELECT username, passwd from user where username = '"+user+"'").Scan(&username, &passwd)
	if err != nil {
		log.Println("get db result failed ", err)
		return nil, err
	} else {
		log.Println("get db result ", username, passwd)
		return []byte(passwd), nil
	}
}
func init_server(ipaddr string) error {
	mysql_handle, err := sql.Open("mysql", *mysql_dsn)
	if err != nil {
		log.Println("connect mysql err ", err)
		return err
	}
	var ip_address string
	err = mysql_handle.QueryRow("SELECT host_ip from user where host_ip = '" + ipaddr + "'").Scan(&ip_address)
	if err != nil {
		sql_cmd := fmt.Sprintf("insert into user (host_ip,cpus,endtime) VALUES ('%s',%d,now())", ipaddr, runtime.NumCPU())
		_, err := mysql_handle.Exec(sql_cmd)
		return err
	} else {
		log.Println(" host ip has exsist ", err)
		sql_cmd := fmt.Sprintf("update user set cpus=%d where host_ip='%s'", runtime.NumCPU(), ipaddr)
		_, err := mysql_handle.Exec(sql_cmd)
		return err
	}
}
func expire_host(address string) error {
	mysql_handle, err := sql.Open("mysql", *mysql_dsn)
	if err != nil {
		log.Println("connect mysql err ", err)
		return err
	}
	sql_str := fmt.Sprintf("update user set username='',ower='',passwd='' where host_ip='%s'", address)
	if _, err = mysql_handle.Exec(sql_str); err != nil {
		return err
	}
	return nil
}
