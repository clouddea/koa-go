package dao

const TABLE_URL = "url"
const TABLE_URL_VERSION = 1
const TABLE_URL_CREATE_SQL = `
	create table url (
	    id     integer PRIMARY KEY AUTOINCREMENT,
		title  text,
		icon   text,   -- url的图标，优先本地获取，如果本地无法获取则尝试云端获取。为空表示获取失败，则前端显示默认
		url    text,
		tags   text,   -- 暂不使用
		type   int,    -- 0, url, 1，目录
		color  int,    -- 暂不使用
		rank   bigint, -- 用于排序
		parent bigint,
		owner  bigint
	)
`

func URL_Create(url URL) {

}
