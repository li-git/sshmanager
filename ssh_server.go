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
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/kr/pty"
	"golang.org/x/crypto/ssh"
)

var (
	hostPrivateKeySigner ssh.Signer
	default_key_str      string
	mysql_dsn            *string
	ssh_port             *string
	keyPath              *string
	encryt_str           *string
	server_sshs          map[*ssh.ServerConn]bool
	local_ip             string
	mutex                sync.Mutex
)

func init() {
	keyPath = flag.String("key", "/root/.ssh/key_rsa", " ssh key path")
	mysql_dsn = flag.String("dsn", "root:ysDn5tOSmJ8=@tcp(10.100.125.17:3306)/logan_test?charset=utf8", "mysql dsn")
	ssh_port = flag.String("sshport", "2222", "sshport")
	encryt_str = flag.String("encryt_str", "", "encryt_str")
	server_sshs = make(map[*ssh.ServerConn]bool)
	default_key_str = `-----BEGIN RSA PRIVATE KEY-----
MIIEogIBAAKCAQEAyUZpKSHo7GFhjbnup+KyjO6FCNgmGo9dI37/3fFDGmYwwR3t
ZeKzKoeolVciG+ezJc3m63C1Yo/2agB6A5OHCskAdB4kcl6b0OyB7ZOydngWxQEn
aTHiFRLPifcUyjABDBOg/utHyqLhEZg8dGikHaYw2VI0hyphWT6QaJ7R8hXoBBdy
+fAADiqGf8OamBdVLixJCjSl0rAAs52sP46H9ct2gzMWqbknLDesF9mcqMiG7ngk
3x1HfU4T6tIkK3tLWBEStTMNLlszYlQGVC+PX/QuzEY8KiSvtCdVMjbovofkPsqg
hSKeFLt6y/nNDx+KYinZpakb8g6YX8SylP8fFwIDAQABAoIBABfYM5Uf40w9rSTx
JgjVnnl7uF733GvBGDOgVAejEWQLPzNhrEIpvTgIojwu+md754lO/1BdJd/rVjHw
pIPP8mugrGEVQRQbiTITEsFmgfnu+COWo2ie9D2y4Mtjbh8V2MnpeWU50mN7MFa5
RlA0JV0t1xOn3Xk12BqOguUiC5U2NSVff5aWBCPN8qNH8RQ9/wr9qnM91RYCwemT
+VVg9NSXsRwLorq0tlFWAU0IyP8nhOOOJJFhWSjm6mkPohBWXs/SQFmWjGZoNhKk
MEW+3HhuUmoY1/mbRMW0XCxv5n+4UYbGpym9cOZmudN9C8Of8PSCZ7IUoSS/aTh8
/xoGiUECgYEA+UbLag5WaeBJy+UFJwkveo0BaurVJlYUjLhFFGFRSWQuMlPCUnvA
qBW0MRupMf6mx7FutEf+sxUwu5OFBFitdFaTdC9I000pWLNUgHHISsdqQMCNl08R
gbSCB3o9H6g9KRtmWoAIDmcyDbcI82OgWSJwRG3/F6XOYHKAiH6Jyc8CgYEAzrQs
4H41p8rkE9qgPI2/yGVlxnvzd/lhjjeO1hNylltj1nX6PKaJiuP9IO91deCV8P3k
o2JMyiTZDF9pPvNd5ISv3S0wU5Ed7ENM+otAjQb4n9IcBXNvB3RU65DNPdKrwvMd
HtXd77MWrHLAmlcYsbIYZ9S8PkjNrCA8IfFX0DkCgYAwt205pOOufW7usit3nYvx
32zPgGV3wIrzlW+qs/o25aVBoKzxgc39C4DTuBww8RuXG04PXaKhTRrhDcuJNetw
ORtIMZWB9iqGc0WodJQ4SRCy5u7FC2bYenaPD4yyiyaoyfoO5catSe22UHcnWekU
gm5+cSDRdk4G+1mzU0eKcQKBgFs2tvb5usOojK0WNM+D3bWYySilWfL/YUVYzvc4
7b/b5FqnBR3uf5OCuBjoknTJ/mCyKUrP/gLV79G96LuWuUA2LUT0w/acew/fQwDs
ojeZc+1S0nq1TbGEbFTnOSqm5JTKo3cP+TflV4QRv1xcQtFnPc3T2p3BksD6GI8B
6TZBAoGAAMtt6F/1rfpCxpAPMed7EFxC37U3bEcS5hPeZtyc6vNrtDEAoqCdZ2Ly
dKeF4r2rMQzVy6AaWdvX1XQPEUXvGF0cjDjrGjwfgZirU56uQ+zbLbc+AfREp/sX
+1gWTj+iq0t7TqeV8vPUQpCpq1xRdhgfXd9gaUoQ5mNui2N7vBI=
-----END RSA PRIVATE KEY-----`
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
		for {
			endtime, err := get_endtime(local_ip)
			if err != nil {
				log.Println("timer refresh failed, ", err)
				break
			} else {
				endstamp, err := time.Parse("2006-01-02 15:04:05", string(endtime))
				if err == nil {
					now, err := get_time()
					if err == nil {
						nowstamp, err := time.Parse("2006-01-02 15:04:05", now)
						if err == nil && nowstamp.Unix() > endstamp.Unix() {
							break
						}
					}
				}
			}
			time.Sleep(time.Duration(10) * time.Second)
		}
		del_ssh_all()
		expire_host(local_ip)
		log.Println("host expired ", local_ip)
		time.Sleep(time.Duration(10) * time.Second)
	}
}
func main() {
	flag.Parse()
	hostPrivateKey, err := ioutil.ReadFile(*keyPath)
	if err != nil {
		log.Println("not found ssh key , use defalt ")
		hostPrivateKey = []byte(default_key_str)
	}
	hostPrivateKeySigner, err = ssh.ParsePrivateKey(hostPrivateKey)
	if err != nil {
		panic(err)
	}
	if len(*encryt_str) > 0 {
		log.Println("encrypte result ", private_encode(*encryt_str))
		os.Exit(0)
	}
	reg := regexp.MustCompile(":(.*)@")
	pass_crypt := reg.FindStringSubmatch(*mysql_dsn)[1]
	*mysql_dsn = strings.Replace(*mysql_dsn, pass_crypt, private_decode(pass_crypt), 1)

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
