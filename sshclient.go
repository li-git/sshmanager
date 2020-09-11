package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
)

type ClientConfig struct {
	Host     string
	Port     int64
	Username string
	Password string
	Client   *ssh.Client
}

var (
	pass *string
)

func init() {
	pass = flag.String("pass", "1234", "password")
}
func (cliConf *ClientConfig) createClient(host string, port int64, username, password string) {
	var (
		client *ssh.Client
		err    error
	)
	cliConf.Host = host
	cliConf.Port = port
	cliConf.Username = username
	cliConf.Password = password
	cliConf.Port = port
	config := ssh.ClientConfig{
		User: cliConf.Username,
		Auth: []ssh.AuthMethod{ssh.Password(password)},
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
		Timeout: 10 * time.Second,
	}
	addr := fmt.Sprintf("%s:%d", cliConf.Host, cliConf.Port)

	if client, err = ssh.Dial("tcp", addr, &config); err != nil {
		log.Fatalln(err)
	}
	cliConf.Client = client
}

func RunShell(client *ssh.Client, shell string) string {
	var (
		session *ssh.Session
		err     error
		output  []byte
	)
	if session, err = client.NewSession(); err != nil {
		log.Fatalln(err)
	}

	if output, err = session.CombinedOutput(shell); err != nil {
		log.Fatalln(err)
	}
	return string(output)
}
func sshSession(client *ssh.Client) {
	var session *ssh.Session
	var err error
	if session, err = client.NewSession(); err != nil {
		log.Fatalln(err)
	}
	session.Stdout = os.Stdout
	session.Stderr = os.Stderr
	session.Stdin = os.Stdin

	termWidth, termHeight, err := terminal.GetSize(int(os.Stdin.Fd()))
	modes := ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}
	if err := session.RequestPty("xterm", termHeight, termWidth, modes); err != nil {
		log.Fatal(err)
	}
	session.Run("bash")
}
/*
func main() {
	flag.Parse()
	cliConf := new(ClientConfig)
	cliConf.createClient("10.100.123.147", 22, "root", *pass)
	//fmt.Println(RunShell(cliConf.Client, "cd /home; ls -l"))
	sshSession(cliConf.Client)
}
*/
