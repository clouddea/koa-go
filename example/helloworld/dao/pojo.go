package dao

type User struct {
	Id       int64
	Nickname string // 显示的名称
	Wechat   string // 绑定的微信
	Email    string // 邮箱认证登录
	Visa     string // 签证，一个证书文件
	Role     User_Role_Type
}

type URL struct {
	Id     int64
	Title  string
	Icon   string // url的图标，优先本地获取，如果本地无法获取则尝试云端获取。为空表示获取失败，则前端显示默认
	Url    string
	Tags   string   // 暂不使用
	Type   URL_Type // 0, url, 1，目录
	Color  int      // 暂不使用
	Rank   int64    //用于排序
	Parent int64
	Owner  int64
}

type URL_Type = int

const URL_TYPE_URL = 0
const URL_TYPE_DIR = 1

type User_Role_Type = int

const USER_ROLE_TYPE_USER = 0
const USER_ROLE_TYPE_ADMIN = 1

type Response struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data,omitempty"`
}

func ResponseUnknow() Response {
	return Response{
		Code: RESPONSE_CODE_UNKNOW,
		Msg:  "未知的错误",
		Data: nil,
	}
}

func ResponseError(msg string) Response {
	return Response{
		Code: RESPONSE_CODE_ERROR,
		Msg:  msg,
		Data: nil,
	}
}

func ResponseCode(code int, msg string) Response {
	return Response{
		Code: code,
		Msg:  msg,
		Data: nil,
	}
}

func ResponseSuccess() Response {
	return Response{
		Code: RESPONSE_CODE_SUCCESS,
		Msg:  "操作成功",
		Data: nil,
	}
}

func ResponseDataSuccess(data any) Response {
	return Response{
		Code: RESPONSE_CODE_SUCCESS,
		Msg:  "操作成功",
		Data: data,
	}
}

const RESPONSE_CODE_SUCCESS = 0
const RESPONSE_CODE_UNKNOW = -1
const RESPONSE_CODE_ERROR = -2
