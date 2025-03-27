# Simple SCP

一个使用Go语言实现的简单SCP（Secure Copy Protocol）工具，支持文件和目录的安全传输。

## 主要功能

- 支持文件上传和下载
- 支持目录递归传输
- 支持密码和SSH密钥认证
- 实时传输进度显示
- 支持相对路径和绝对路径
- 支持 `~` 路径展开
- 自动处理目标路径是否为目录的情况

## 实现原理

- 基于Go语言的 `golang.org/x/crypto/ssh` 包实现SSH连接
- 使用SCP协议进行文件传输
- 使用 `progressbar` 库实现进度条显示
- 支持多种认证方式：
  - 密码认证
  - SSH密钥认证（默认使用 `~/.ssh/id_rsa`）

## 安装

```bash
go install github.com/Lyrichu/simple_scp/cmd/simple_scp@latest
```

## 使用方法

### 命令行参数

```bash
simple_scp -source <source> -dest <dest> -user <user> [-password <password>] [-port <port>] [-r]
```

参数说明：
- `-source`, `-s`: 源路径
- `-dest`, `-d`: 目标路径
- `-user`, `-u`: SSH用户名
- `-password`, `-pw`: SSH密码（可选，默认使用SSH密钥）
- `-port`, `-p`: SSH端口（可选，默认22）
- `-r`: 递归传输目录（可选）

### 使用示例

1. 上传本地文件到远程服务器：
```bash
simple_scp -s ./local_file.txt -d user@remote:/path/to/dest -u user
```

2. 从远程服务器下载文件：
```bash
simple_scp -s user@remote:/path/to/file.txt -d ./local_path -u user
```

3. 递归上传目录：
```bash
simple_scp -s ./local_dir -d user@remote:/path/to/dest -u user -r
```

4. 使用密码认证：
```bash
simple_scp -s ./local_file.txt -d user@remote:/path/to/dest -u user -pw your_password
```

5. 指定自定义端口：
```bash
simple_scp -s ./local_file.txt -d user@remote:/path/to/dest -u user -p 2222
```

## 特点

1. 智能路径处理：
   - 自动处理相对路径和绝对路径
   - 支持 `~` 展开为用户主目录
   - 自动处理目标为目录的情况

2. 用户友好：
   - 实时显示传输进度
   - 支持简写参数形式
   - 清晰的错误提示

3. 安全性：
   - 支持SSH密钥认证
   - 支持密码认证
   - 基于SSH协议保证传输安全

## 注意事项

1. 使用密码认证时，密码将在命令行中明文显示，建议使用SSH密钥认证
2. 递归传输目录时需要使用 `-r` 参数
3. 默认使用 `~/.ssh/id_rsa` 作为SSH密钥