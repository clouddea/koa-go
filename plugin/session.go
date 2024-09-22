package plugin

import (
	"github.com/clouddea/koa-go/koa"
	"github.com/clouddea/koa-go/koa/util"
	"github.com/google/uuid"
	"net/http"
)

type Session struct {
	Id  string
	Val map[string]any
}

/*
*
最简Session实现
@param maxAge 超时时间(s)
*/
func NewSession(capacity int, maxAge int) koa.PluginMultiArg {
	var sessionStore = util.NewLRUCache[string](capacity)
	return func(context *koa.Context, next func()) {
		sessionKey := "KOA_SESSION_ID"
		cookie, err := context.Req.Cookie(sessionKey)
		if err == nil && cookie != nil {
			sessionId := cookie.Value
			// log.Println(context.Req.URL.Path + "==>" + sessionId)
			if sessionVal, ok := sessionStore.Get(sessionId); ok {
				context.State["session"] = sessionVal.(*Session)
				next()
				return
			}
		}
		// 生成新的Session
		// UUID会重复吗？ https://blog.csdn.net/u012760435/article/details/122304214
		// https://huanglianjing.com/posts/uuidgo%E9%80%9A%E7%94%A8%E5%94%AF%E4%B8%80%E6%A0%87%E8%AF%86%E7%AC%A6%E7%94%9F%E6%88%90/
		tryCount := 5
		var sessionId = ""
		for tryCount > 0 {
			sessionId = uuid.NewString()
			if _, ok := sessionStore.Get(sessionId); !ok {
				session := &Session{
					Id: sessionId, Val: make(map[string]any),
				}
				sessionStore.Put(sessionId, session)
				context.State["session"] = session
				break
			}
			tryCount--
		}
		if tryCount == 0 {
			panic("can not create uuid for session")
		}
		context.Response.SetCookie(&http.Cookie{
			Name:   sessionKey,
			Value:  sessionId,
			Path:   "/",
			MaxAge: maxAge,
		})
		next()
	}
}
