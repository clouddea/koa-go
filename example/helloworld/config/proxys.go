package config

var Proxys = map[string][]string{
	"/baidu":  []string{"https://www.baidu.com$extra"},
	"/github": []string{"https://github.com$extra"},
}
