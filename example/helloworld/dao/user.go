package dao

import (
	"database/sql"
	"github.com/clouddea/koa-go/koa/util"
)

const TABLE_USER = "user"
const TABLE_USER_VERSION = 3
const TABLE_USER_CREATE_SQL = `
	create table user (
	    id       integer PRIMARY KEY AUTOINCREMENT,
		nickname text, -- 显示的名称
		wechat   text, -- 绑定的微信
		email    text, -- 邮箱认证登录
		visa     text, -- 签证，一个证书文件
		role     int
	)
`

func User_Create(db *sql.DB, user User) int64 {
	result, err := db.Exec(`
		insert into user(
		    nickname,
		    wechat,
		    email,
		    visa,
		    role
		) values (?, ?, ?, ?, ?)
	`, user.Nickname, user.Wechat, user.Email, user.Visa, user.Role)
	util.Assert(err, "create user error")
	insertedId, _ := result.LastInsertId()
	return insertedId
}

func User_Update(db *sql.DB, user User) {
	_, err := db.Exec(`
		update user
		set
		    nickname = ?,
		    wechat = ?,
		    email = ?,
		    visa = ?,
		    role = ?
		where
		    id = ?
	`, user.Nickname, user.Wechat, user.Email, user.Visa, user.Role, user.Id)
	util.Assert(err, "update user error")
}

func User_Query(db *sql.DB, userId int64) (User, bool) {
	rows, err := db.Query(`
		select
		    *
		from user 
		where 
		    id = ?
	`, userId)
	defer rows.Close()
	if !rows.Next() {
		return User{}, false
	}
	user := User{}
	err = rows.Scan(&user.Id, &user.Nickname, &user.Wechat, &user.Email, &user.Visa, &user.Role)
	util.Assert(err, "read user info error")
	return user, true
}
