package dao

import (
	"database/sql"
	"fmt"
	"sync"
)

const TABLE_TEST = "test"
const TABLE_TEST_VERSION = 2
const TABLE_TEST_CREATE_SQL = `
	create table test (
	    id int PRIMARY KEY,
	    name text
	)
`

var prepare *sql.Stmt
var id int = 0
var mutex sync.Mutex

func Test_DAO_Add(db *sql.DB) error {
	if prepare == nil {
		stmt, err := db.Prepare(fmt.Sprintf("insert into %v values(?, 'demo')", TABLE_TEST))
		if err != nil {
			return err
		}
		prepare = stmt
	}
	mutex.Lock()
	id++
	mutex.Unlock()
	_, err := prepare.Exec(id)
	return err
}
