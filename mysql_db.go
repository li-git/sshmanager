package main

import (
	"database/sql"
	"log"

	_ "github.com/Go-SQL-Driver/MySQL"
)

func get_pass(user string) ([]byte, error) {
	mysql_handle, err := sql.Open("mysql", mysql_dsn)
	if err != nil {
		log.Println("===>connect mysql err ", err)
		return nil, err
	}
	var username, passwd string
	err = mysql_handle.QueryRow("SELECT username, passwd from user where username = '"+user+"'").Scan(&username, &passwd)
	if err != nil {
		log.Println("===>get db result failed ", err)
		return nil, err
	} else {
		log.Println("===>get db result ", username, passwd)
		return []byte(passwd), nil
	}
}
