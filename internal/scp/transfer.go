package scp

import (
	"fmt"
	"github.com/schollz/progressbar/v3"
	"golang.org/x/crypto/ssh"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func (c *Client) Transfer(source, destination string) error {
	// 解析源和目标路径
	isLocalToRemote := !strings.Contains(source, ":")
	var host, path string

	if isLocalToRemote {
		parts := strings.SplitN(destination, ":", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid destination format")
		}
		host, path = parts[0], parts[1]
	} else {
		parts := strings.SplitN(source, ":", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid source format")
		}
		host, path = parts[0], parts[1]
	}

	// 处理 ~ 路径
	if strings.HasPrefix(path, "~") {
		path = strings.Replace(path, "~", "/home/"+c.config.User, 1)
	}

	// 如果路径是相对路径，转换为绝对路径
	if !filepath.IsAbs(path) {
		path = filepath.Join("/home", c.config.User, path)
	}

	// 从 host 中提取真实主机名（去除可能存在的用户名）
	if idx := strings.Index(host, "@"); idx != -1 {
		host = host[idx+1:]
	}

	// 建立 SSH 连接
	client, err := c.connect(host)
	if err != nil {
		return fmt.Errorf("failed to connect: %v", err)
	}
	defer client.Close()

	if isLocalToRemote {
		// 检查源路径是否为目录
		stat, err := os.Stat(source)
		if err != nil {
			return err
		}
		if stat.IsDir() {
			if !c.recursive {
				return fmt.Errorf("source is a directory, please use -r flag for recursive transfer")
			}
			// 检查远程目标路径是否存在
			session, err := client.NewSession()
			if err != nil {
				return err
			}
			output, err := session.Output(fmt.Sprintf("test -d %s && echo 'EXISTS' || echo 'NOT_EXISTS'", path))
			session.Close()
			if err != nil {
				return err
			}

			exists := strings.TrimSpace(string(output)) == "EXISTS"
			if exists {
				// 目标目录存在，在其下创建同名目录
				targetPath := filepath.Join(path, filepath.Base(source))
				return c.uploadDirectory(client, source, targetPath)
			} else {
				// 目标目录不存在，直接使用目标路径（重命名）
				return c.uploadDirectory(client, source, path)
			}
		}
		fmt.Printf("Uploading file: %s\n", source)
		return c.uploadFile(client, source, path)
	} else {
		// 检查远程路径是否为目录
		session, err := client.NewSession()
		if err != nil {
			return err
		}
		output, err := session.Output(fmt.Sprintf("test -d %s && echo 'DIR' || echo 'FILE'", path))
		session.Close()
		if err != nil {
			return err
		}

		isDir := strings.TrimSpace(string(output)) == "DIR"
		if isDir {
			if !c.recursive {
				return fmt.Errorf("source is a directory, please use -r flag for recursive transfer")
			}
			// 检查目标路径是否存在
			destInfo, err := os.Stat(destination)
			if err == nil && destInfo.IsDir() {
				// 目标目录存在，在其下创建同名目录
				targetPath := filepath.Join(destination, filepath.Base(path))
				return c.downloadDirectory(client, path, targetPath)
			} else {
				// 目标目录不存在，直接使用目标路径（重命名）
				return c.downloadDirectory(client, path, destination)
			}
		}
		fmt.Printf("Downloading file: %s\n", path)
		// 处理目标路径
		if destination == "." {
			destination = filepath.Base(path)
		} else {
			destInfo, err := os.Stat(destination)
			if err == nil && destInfo.IsDir() {
				destination = filepath.Join(destination, filepath.Base(path))
			}
		}
		return c.downloadFile(client, path, destination)
	}
}

func (c *Client) uploadFile(client *ssh.Client, localPath, remotePath string) error {
	session, err := client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	file, err := os.Open(localPath)
	if err != nil {
		return err
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return err
	}

	fmt.Printf("Uploading file: %s (%d bytes)\n", filepath.Base(localPath), stat.Size())
	bar := progressbar.DefaultBytes(
		stat.Size(),
		fmt.Sprintf("uploading %s", filepath.Base(localPath)),
	)

	w, err := session.StdinPipe()
	if err != nil {
		return err
	}

	if err := session.Start(fmt.Sprintf("scp -t %s", remotePath)); err != nil {
		return err
	}

	fmt.Fprintf(w, "C0644 %d %s\n", stat.Size(), filepath.Base(localPath))

	_, err = io.Copy(io.MultiWriter(w, bar), file)
	if err != nil {
		return err
	}

	fmt.Fprint(w, "\x00")
	w.Close()

	return session.Wait()
}

func (c *Client) downloadFile(client *ssh.Client, remotePath, localPath string) error {
	session, err := client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	// 获取输出和输入管道
	stdout, err := session.StdoutPipe()
	if err != nil {
		return err
	}
	stdin, err := session.StdinPipe()
	if err != nil {
		return err
	}

	// 创建本地文件
	file, err := os.Create(localPath)
	if err != nil {
		return err
	}
	defer file.Close()

	// 启动远程 scp 进程
	if err := session.Start(fmt.Sprintf("scp -f %s", remotePath)); err != nil {
		return err
	}

	// 发送确认信号
	stdin.Write([]byte{0})

	// 读取文件信息
	buffer := make([]byte, 1024)
	n, err := stdout.Read(buffer)
	if err != nil {
		return err
	}

	// 解析文件信息（格式：C0644 filesize filename）
	fileInfo := string(buffer[:n])
	var mode, fileSize int64
	var filename string
	if _, err := fmt.Sscanf(fileInfo, "C%o %d %s", &mode, &fileSize, &filename); err != nil {
		return fmt.Errorf("failed to parse file info: %v, raw info: %s", err, fileInfo)
	}

	fmt.Printf("Downloading file: %s (%d bytes)\n", filename, fileSize)

	// 发送确认信号
	stdin.Write([]byte{0})

	// 创建进度条
	bar := progressbar.DefaultBytes(
		fileSize,
		fmt.Sprintf("downloading %s", filename),
	)

	// 创建一个有限的 Reader，只读取文件大小的内容
	limitReader := io.LimitReader(stdout, fileSize)

	// 复制文件内容
	_, err = io.Copy(io.MultiWriter(file, bar), limitReader)
	if err != nil {
		return err
	}

	// 读取并确认结束信号
	endBuffer := make([]byte, 1)
	_, err = stdout.Read(endBuffer)
	if err != nil && err != io.EOF {
		return err
	}

	// 发送最后的确认信号
	stdin.Write([]byte{0})

	// 等待命令完成
	return session.Wait()
}

func (c *Client) uploadDirectory(client *ssh.Client, localPath, remotePath string) error {
	fmt.Printf("Creating directory: %s\n", filepath.Base(localPath))

	// 创建远程目录
	session, err := client.NewSession()
	if err != nil {
		return err
	}
	session.Run(fmt.Sprintf("mkdir -p %s", remotePath))
	session.Close()

	files, err := os.ReadDir(localPath)
	if err != nil {
		return err
	}

	for _, file := range files {
		localFilePath := filepath.Join(localPath, file.Name())
		remoteFilePath := filepath.Join(remotePath, file.Name())

		if file.IsDir() {
			if c.recursive {
				fmt.Printf("Processing directory: %s\n", file.Name())
				if err := c.uploadDirectory(client, localFilePath, remoteFilePath); err != nil {
					return err
				}
			}
			continue
		}

		if err := c.uploadFile(client, localFilePath, remoteFilePath); err != nil {
			return err
		}
	}
	return nil
}

func (c *Client) downloadDirectory(client *ssh.Client, remotePath, localPath string) error {
	fmt.Printf("Creating directory: %s\n", filepath.Base(localPath))

	if err := os.MkdirAll(localPath, 0755); err != nil {
		return err
	}

	session, err := client.NewSession()
	if err != nil {
		return err
	}
	output, err := session.Output(fmt.Sprintf("ls -la %s", remotePath))
	session.Close()
	if err != nil {
		return err
	}

	files := strings.Split(string(output), "\n")
	for _, file := range files[1:] {
		if file == "" {
			continue
		}
		parts := strings.Fields(file)
		if len(parts) < 9 {
			continue
		}
		filename := parts[8]
		if filename == "." || filename == ".." {
			continue
		}

		remoteFilePath := filepath.Join(remotePath, filename)
		localFilePath := filepath.Join(localPath, filename)

		if strings.HasPrefix(parts[0], "d") {
			if c.recursive {
				fmt.Printf("Processing directory: %s\n", filename)
				if err := c.downloadDirectory(client, remoteFilePath, localFilePath); err != nil {
					return err
				}
			}
			continue
		}

		if err := c.downloadFile(client, remoteFilePath, localFilePath); err != nil {
			return err
		}
	}
	return nil
}
