package koa

import "fmt"

func Assert(err error, msg string) {
	if err != nil {
		panic(fmt.Sprintf("%v\n%v\n", msg, err))
	}
}
