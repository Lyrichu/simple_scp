package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"simple_scp/internal/scp"
	"strings"
)

func main() {
	var (
		source      string
		destination string
		user        string
		password    string
		port        int
		recursive   bool
	)

	// 添加完整形式和缩写形式的参数
	flag.StringVar(&source, "source", "", "source path")
	flag.StringVar(&source, "s", "", "source path (shorthand)")
	flag.StringVar(&destination, "dest", "", "destination path")
	flag.StringVar(&destination, "d", "", "destination path (shorthand)")
	flag.StringVar(&user, "user", "", "ssh user")
	flag.StringVar(&user, "u", "", "ssh user (shorthand)")
	flag.StringVar(&password, "password", "", "ssh password")
	flag.StringVar(&password, "pw", "", "ssh password (shorthand)")
	flag.IntVar(&port, "port", 22, "ssh port")
	flag.IntVar(&port, "p", 22, "ssh port (shorthand)")
	flag.BoolVar(&recursive, "r", false, "recursive copy")
	flag.Parse()

	// 如果没有提供用户名，尝试从源或目标地址中提取
	if user == "" {
		if strings.Contains(source, "@") {
			parts := strings.SplitN(source, "@", 2)
			user = parts[0]
		} else if strings.Contains(destination, "@") {
			parts := strings.SplitN(destination, "@", 2)
			user = parts[0]
		}
	}

	if source == "" || destination == "" || user == "" {
		fmt.Println("Usage: simple_scp -source <source> -dest <dest> -user <user> [-password <password>] [-port <port>]")
		os.Exit(1)
	}

	client := scp.NewClient(user, password, port, recursive)
	if err := client.Transfer(source, destination); err != nil {
		log.Fatal(err)
	}
}
