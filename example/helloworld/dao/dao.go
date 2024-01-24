package dao

import (
	"database/sql"
	"fmt"
	"github.com/clouddea/koa-go/koa"
)

const _TABLE_META_EXIST_QUERY = `
	select * from sqlite_master where tbl_name = 'table_version'
`

const _TABLE_CREATE_SQL = `
	create table table_version
	(
		tb_name  varchar(32),
		tb_version UNSIGNED INTEGER NOT NULL,
		comment varchar(128)
	)
`
const _TABLE_VERSION_INSERT_SQL = `
    insert into table_version values(
        '%v',
        %v,
    	'a talbe for test'
    )
`

const _TABLE_VERSION_QUERY_SQL = `
	select * from table_version 
	where tb_name = '%s' and tb_version = %v 
`

const _TABLE_DROP_SQL = `
	drop table if exists %v
`

const _TABLE_VERSION_DELETE_SQL = `
	delete from table_version
	where tb_name = '%v'
`

func DAOInit(db *sql.DB) {
	// 初始配置
	_, err := db.Exec("PRAGMA synchronous = OFF")
	koa.Assert(err, "init dao error: config sqlite error")
	// 初始化元数据表
	tx, err := db.Begin()
	koa.Assert(err, "init dao error: can not open transcantion")
	rows, err := tx.Query(_TABLE_META_EXIST_QUERY)
	koa.Assert(err, "init dao error: can not query table 'table_version' ")
	if !rows.Next() {
		_, err := tx.Exec(_TABLE_CREATE_SQL)
		koa.Assert(err, "init dao error: create table 'table_version' error")
	}
	koa.Assert(tx.Commit(), "init dao error: can not commit transaction ")
	// 初始化各个表
	ensureTalbe(db, TABLE_TEST, TABLE_TEST_VERSION, TABLE_TEST_CREATE_SQL)
	ensureTalbe(db, TABLE_USER, TABLE_USER_VERSION, TABLE_USER_CREATE_SQL)
	ensureTalbe(db, TABLE_URL, TABLE_URL_VERSION, TABLE_URL_CREATE_SQL)
}

func ensureTalbe(db *sql.DB, name string, version int, sql string) {
	tx, err := db.Begin()
	koa.Assert(err, "ensure table error: can not open transcantion")
	rows, err := tx.Query(fmt.Sprintf(_TABLE_VERSION_QUERY_SQL, name, version))
	koa.Assert(err, "ensure table error: can not query table version")
	if !rows.Next() {
		_, err := tx.Exec(fmt.Sprintf(_TABLE_DROP_SQL, name))
		koa.Assert(err, "ensure table error: can not drop old version table")
		_, err = tx.Exec(fmt.Sprintf(_TABLE_VERSION_DELETE_SQL, name))
		koa.Assert(err, "ensure table error: can not drop old version table info")
		_, err = tx.Exec(sql)
		koa.Assert(err, "ensure table error: can not create table")
		_, err = tx.Exec(fmt.Sprintf(_TABLE_VERSION_INSERT_SQL, name, version))
		koa.Assert(err, "ensure table error: can not create table version")
	}
	koa.Assert(tx.Commit(), "ensure table error: can not commit transaction ")
}
