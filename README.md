# 黄桃罐头QQKEY

一款基于 **Go + Fyne** 开发的轻量级桌面工具，帮助用户快速登录、管理 QQ 账号，并一键跳转至腾讯系常用站点（如 Qzone、QQ群、邮箱等），提升日常操作效率。

### ❗说明
- ⚠该程序仅用于学习交流，请勿违法使用，所产生的后果作者不负法律责任
- 客户端源码在 ./client/createClient 内
- 网络上还有很多公开的qqkey api懒得加
  
## 🌟 原理
- 网络上开源的QQ漏洞访问接口
- 内存扫描特征码，避开“修改host文件，抓包方式防止QQkey被获取”

## 🌟 核心功能
- ✅ **一键跳转腾讯站点**：
  - QQ空间
  - QQ群管理
  - QQ邮箱
- ✅ **控制**：
  - 查看/发送公告
  - 查看/转发群精华
  - 查看/删除群文件
  - 修改个人资料，查看群列表

- ✅ **通过代理进入几乎所有可用QQ快捷登录的站点**：腾讯视频，qq音乐，腾讯云游戏
- ✅ **配置持久化**：监听端口、上线地址等配置自动保存至本地


## 🛠️ 技术栈

- **语言**：Go 1.21+
- **UI 框架**：[Fyne](https://fyne.io/)（Go GUI 框架）
- **网络中间件**:[Sunnynet] (github.com/qtgolang/SunnyNet/SunnyNet)(用于快捷登录)

### 安装Fyne ClI

```bash
go install fyne.io/fyne/v2/cmd/fyne@latest
```
### 安装TDM-GCC(https://jmeubank.github.io/tdm-gcc/)
- 配置自行百度

### 构建项目
```bash
# 安装依赖
go mod tidy

set GOOS=windows
#设置64/32位
set GOARCH=amd64

set CGO_ENABLED=1

#构建项目
go build -ldflags="-s -w" -o HTGT.exe .
```
- 或直接使用go build.bat 或 fyne build.bat
- 客户端(./client/createClient)同上，记得将生成后的客户端改名loader.exe 放入 ./client 目录

