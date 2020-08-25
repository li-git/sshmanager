package main

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os/exec"
	"strings"
	"sync"

	"github.com/kr/pty"
	"golang.org/x/crypto/ssh"
)

var (
	hostPrivateKeySigner ssh.Signer
	mysql_dsn            string
	ssh_port             string
)

func init() {
	keyPath := "/root/.ssh/key_rsa"
	hostPrivateKey, err := ioutil.ReadFile(keyPath)
	if err != nil {
		panic(err)
	}
	hostPrivateKeySigner, err = ssh.ParsePrivateKey(hostPrivateKey)
	if err != nil {
		panic(err)
	}
	mysql_dsn = "root:Pass_123@tcp(10.100.125.17:3306)/logan_test?charset=utf8"
	ssh_port = "2222"
}
func handleChannel(newChannel ssh.NewChannel) {
	if t := newChannel.ChannelType(); t != "session" {
		newChannel.Reject(ssh.UnknownChannelType, fmt.Sprintf("unknown channel type: %s", t))
		return
	}
	connection, requests, err := newChannel.Accept()
	if err != nil {
		log.Printf("Could not accept channel (%s)", err)
		return
	}
	bash := exec.Command("bash")
	close := func() {
		connection.Close()
		_, err := bash.Process.Wait()
		if err != nil {
			log.Printf("Failed to exit bash (%s)", err)
		}
		log.Printf("Session closed")
	}
	bashf, err := pty.Start(bash)
	if err != nil {
		log.Printf("Could not start pty (%s)", err)
		close()
		return
	}
	var once sync.Once
	go func() {
		io.Copy(connection, bashf)
		once.Do(close)
	}()
	go func() {
		io.Copy(bashf, connection)
		once.Do(close)
	}()
	go func() {
		for req := range requests {
			switch req.Type {
			case "shell":
				if len(req.Payload) == 0 {
					req.Reply(true, nil)
				}
			case "pty-req":
				termLen := req.Payload[3]
				w, h := parseDims(req.Payload[termLen+4:])
				SetWinsize(bashf.Fd(), w, h)
				req.Reply(true, nil)
			case "window-change":
				w, h := parseDims(req.Payload)
				SetWinsize(bashf.Fd(), w, h)
			}
		}
	}()
}
func keyAuth(conn ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
	log.Println(conn.RemoteAddr(), "authenticate with", key.Type(), " User ", conn.User())
	return nil, errors.New("not support ssh key login")
}
func passAuth(conn ssh.ConnMetadata, password []byte) (*ssh.Permissions, error) {
	log.Println("passAuth-->connect user ", conn.User(), " pass ", string(password))
	passwd, err := get_pass(conn.User())
	if err != nil {
		return nil, err
	} else if strings.Compare(string(password), string(passwd)) == 0 {
		return nil, nil
	} else {
		return nil, errors.New("check user failed")
	}
}
func ssh_working(chans <-chan ssh.NewChannel, sshConn *ssh.ServerConn) {
	for newChannel := range chans {
		go handleChannel(newChannel)
	}
	sshConn.Close()
}
func main() {
	config := ssh.ServerConfig{
		//PublicKeyCallback: keyAuth,
		PasswordCallback: passAuth,
	}
	config.AddHostKey(hostPrivateKeySigner)
	socket, err := net.Listen("tcp", ":"+ssh_port)
	if err != nil {
		log.Println("tcp listen failed ", err)
	}
	for {
		conn, err := socket.Accept()
		if err != nil {
			log.Println("tcp accept failed ", err)
		}
		sshConn, chans, reqs, err := ssh.NewServerConn(conn, &config)
		if err != nil {
			log.Println("conect failed ", err)
		} else {
			log.Println("Connection from", sshConn.RemoteAddr())
			go ssh.DiscardRequests(reqs)
			go ssh_working(chans, sshConn)
		}
	}
}
