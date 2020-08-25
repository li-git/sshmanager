package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/kr/pty"
	"golang.org/x/crypto/ssh"
)

var (
	hostPrivateKeySigner ssh.Signer
	mysql_dsn            *string
	ssh_port             *string
	keyPath              *string
	server_sshs          map[*ssh.ServerConn]bool
	local_ip             string
	mutex                sync.Mutex
)

func init() {
	keyPath = flag.String("key", "/root/.ssh/key_rsa", " ssh key path")
	hostPrivateKey, err := ioutil.ReadFile(*keyPath)
	if err != nil {
		panic(err)
	}
	hostPrivateKeySigner, err = ssh.ParsePrivateKey(hostPrivateKey)
	if err != nil {
		panic(err)
	}
	mysql_dsn = flag.String("dsn", "root:Pass_123@tcp(10.100.125.17:3306)/logan_test?charset=utf8", "mysql dsn")
	ssh_port = flag.String("sshport", "2222", "sshport")
	server_sshs = make(map[*ssh.ServerConn]bool)
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
	//sshConn.Close()
	del_ssh_conn(sshConn)
}
func add_ssh_conn(sshConn *ssh.ServerConn) {
	mutex.Lock()
	server_sshs[sshConn] = true
	mutex.Unlock()
}
func del_ssh_conn(sshConn *ssh.ServerConn) {
	sshConn.Close()
	mutex.Lock()
	delete(server_sshs, sshConn)
	mutex.Unlock()
}
func del_ssh_all() {
	mutex.Lock()
	for k := range server_sshs {
		k.Close()
		delete(server_sshs, k)
	}
	mutex.Unlock()
}
func check_timer() {
	for {
		endtime, err := get_endtime(local_ip)
		if err != nil {
			log.Println("timer refresh failed, ", err)
		} else {
			endstamp, err := time.Parse("2006-01-02 15:04:05", string(endtime))
			if err == nil && time.Now().Unix() > endstamp.Unix() {
				del_ssh_all()
				log.Println("host expired ", local_ip)
			}
			//log.Println("===========>", string(endtime))
		}
		time.Sleep(time.Duration(5) * time.Second)
	}
}
func main() {
	flag.Parse()
	sys_ip := get_local_ip()
	if sys_ip != nil {
		log.Println("local ip is ", string(sys_ip))
		local_ip = string(sys_ip)
	} else {
		log.Println("can not get local ip")
		os.Exit(0)
	}
	init_server(string(local_ip))
	go check_timer()
	config := ssh.ServerConfig{
		//PublicKeyCallback: keyAuth,
		PasswordCallback: passAuth,
	}
	config.AddHostKey(hostPrivateKeySigner)
	socket, err := net.Listen("tcp", ":"+*ssh_port)
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
			add_ssh_conn(sshConn)
		}
	}
}
