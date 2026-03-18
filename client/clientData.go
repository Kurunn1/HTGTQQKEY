package client

import (
	_ "embed"
)

//go:embed loader.exe
var LoaderData []byte

//这里面是用来生成客户端的数据,createclient里面为客户端源码
