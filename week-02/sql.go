package main

import (
	"database/sql"
	//"database/sql/driver"
	_ "github.com/go-sql-driver/mysql"
	"github.com/pkg/errors"
)

const (
	sqlUsr = "root"
	sqlPwd = "123456"
)

func initSql() (db *sql.DB, err error) {
	db, err = sql.Open("mysql", sqlUsr+":"+sqlPwd+"@tcp(127.0.0.1:3306)/tc-go?charset=utf8")
	if err != nil {
		err = errors.Wrap(err, "open sql fail")
		return
	}
	err = db.Ping()
	if err != nil {
		err = errors.Wrap(err, "link sql fail")
		return
	}
	return
}

func query(db *sql.DB, sqlStr string) (err error) {
	rows, err := db.Query(sqlStr)
	if err != nil {
		return errors.Wrap(err, "query user fail")
	}
	for rows.Next() {
		var usr user
		err = rows.Scan(&usr.Id, &usr.Name, &usr.status)
		if err != nil {
			switch {
			case err == sql.ErrNoRows:
				err = errors.Wrap(err, "ErrNoRows")
			default:
				err = errors.Wrap(err, "Scan error")
			}
			return
		}
	}
	return
}

type user struct {
	Id     int64
	Name   string
	status int8
}
