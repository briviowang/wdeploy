package runtime

import (
	"os/exec"
	"time"
)

// SSHService ssh上传、执行shell
type SSHService struct {
	config WebDeployConfigItem

	Output string

	inited bool

	sshCommonParams   []string
	scpCommonParams   []string
	forwardRemoteAddr string
}

// SSHHostKey 配置项
type SSHHostKey struct {
	IP  string `json:"ip"`
	Key string `json:"key"`
}

func (c *SSHService) init() {
	if c.inited {
		return
	}

	c.sshCommonParams = []string{
		"ssh",
		"-p " + c.config.ServerPort,
		"-o StrictHostKeyChecking=no",
		c.config.ServerUser + "@" + c.config.ServerHost,
	}
	c.scpCommonParams = []string{
		"scp",
		"-P " + c.config.ServerPort,
		"-o StrictHostKeyChecking=no",
	}
	if !IsEmpty(c.config.SSHPrivateFile) {
		c.sshCommonParams = append(c.sshCommonParams, "-i", c.config.SSHPrivateFile)
		c.scpCommonParams = append(c.scpCommonParams, "-i", c.config.SSHPrivateFile)
	}
	if !IsEmpty(c.config.SSHJumpHost) {
		proxyCommand := "ProxyCommand=ssh " + c.config.SSHJumpHost + " nc -w 1 %h %p"
		c.sshCommonParams = append(c.sshCommonParams, "-o", proxyCommand)
		c.scpCommonParams = append(c.scpCommonParams, "-o", proxyCommand)
	}

	// if !IsEmpty(c.config.SSHProxyCommand) {
	// 	c.sshCommonParams = append(c.sshCommonParams, "-o", c.config.SSHProxyCommand)
	// 	c.scpCommonParams = append(c.scpCommonParams, "-o", c.config.SSHProxyCommand)
	// }
}

// Shell 用法：
//
//	ssh.Shell("pwd")
func (c *SSHService) Shell(code string) {
	c.init()
	Execute(append(c.sshCommonParams, []string{
		"set -e \n" + code,
	}...))

}

// ShellQuiet 静默执行shell
func (c *SSHService) ShellQuiet(code string) string {
	c.init()
	return ExecuteQuiet(append(c.sshCommonParams, []string{
		"set -e \n" + code,
	}...))
}

// ExecuteFile 执行shell文件
func (c *SSHService) ExecuteFile(name string, data map[string]interface{}) string {
	c.init()

	outShellFile := Pwd() + "/.git/deploy/server.sh"

	contentBytes, err := Asset("assets/scripts/" + name + ".sh")
	CheckErr(err)

	content := ParseTemplate(string(contentBytes), data)
	FilePutContents(outShellFile, content)

	respContent := ""

	respContent = ExecuteQuiet(append(c.sshCommonParams, []string{
		outShellFile,
	}...))

	Rm(outShellFile)
	return respContent
}

// Download 用法:
//
//	ssh.Download("/root/tan.tar.gz", "./")
func (c *SSHService) Download(remotePath string, localPath string) {
	c.init()

	Execute(append(c.scpCommonParams, []string{
		"-r",
		"-p",
		c.config.ServerUser + "@" + c.config.ServerHost + ":" + remotePath,
		c.path(localPath),
	}...))

}

// Upload 上传文件
func (c *SSHService) Upload(localPath string, remotePath string) {
	c.init()

	params := append(c.scpCommonParams, []string{
		"-r",
		"-p",
		c.path(localPath),
		c.config.ServerUser + "@" + c.config.ServerHost + ":" + remotePath,
	}...)
	InfoLog(Implode(" ", params))
	Execute(params)
}

func (c *SSHService) path(localPath string) string {
	//如果当前环境是cygwin，则转换localPath
	if IsWindows() {
		localPath = ExecuteQuiet([]string{
			"cygpath",
			"-u",
			localPath,
		})
	}
	return localPath
}

// SetForwardRemoteAddr 设置远程地址、端口
func (c *SSHService) SetForwardRemoteAddr(addr string) {
	c.forwardRemoteAddr = addr
}

// MysqlForward mysql端口转发
func (c *SSHService) MysqlForward() (res *exec.Cmd, p int) {
	c.init()

	var cmd *exec.Cmd
	port := GetRandomPort()

	if IsEmpty(c.forwardRemoteAddr) {
		c.forwardRemoteAddr = "127.0.0.1:3306"
	}

	cmd = ExecuteStart(append(c.sshCommonParams, []string{
		"-N",
		"-S none",
		"-o ControlMaster=no",
		"-o ExitOnForwardFailure=yes",
		"-o ConnectTimeout=10",
		"-o NumberOfPasswordPrompts=3",
		"-o TCPKeepAlive=no",
		"-o ServerAliveInterval=60",
		"-o ServerAliveCountMax=1",
		"-L",
		IntToString(port) + ":" + c.forwardRemoteAddr,
	}...))

	for {
		time.Sleep(500 * time.Millisecond)
		if !IsPortCanUse(port) {
			return cmd, port
		}
	}
}

// NewSSHService 实例化SSHService
func NewSSHService(config WebDeployConfigItem) SSHService {
	return SSHService{
		config: config,
	}
}
func initSSHService() SSHService {
	return NewSSHService(*currentDeployConfig)
}
