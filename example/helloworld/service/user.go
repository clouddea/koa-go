package service

import (
	"database/sql"
	"github.com/clouddea/koa-go/example/helloworld/dao"
)

func User_Register_Service(db *sql.DB, user dao.User) {
	dao.User_Create(db, user)
}

func User_Update_Service(db *sql.DB, user dao.User) {
	dao.User_Update(db, user)
}
