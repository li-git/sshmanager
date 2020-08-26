package main

/*
CREATE TABLE `user` (
  `host_ip` varchar(255) NOT NULL,
  `cpus` int(32) NOT NULL,
  `username` varchar(255) NOT NULL DEFAULT '',
  `passwd` varchar(255) NOT NULL DEFAULT '',
  `ower` varchar(255) NOT NULL DEFAULT '',
  `endtime` datetime NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
*/
import (
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
)

var (
	httpserver *string
	mysql_dsn  *string
)

func init() {
	httpserver = flag.String("httpserver", "0.0.0.0:88", "httpserver addr")
	mysql_dsn = flag.String("dsn", "root:Pass_123@tcp(10.100.125.17:3306)/logan_test?charset=utf8", "mysql dsn")
}
func main() {
	flag.Parse()
	go http_server_run(*httpserver)
	c := make(chan bool)
	<-c
}
func http_server_run(httpserver string) {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		t, err := template.ParseFiles("index.html")
		if err != nil {
			log.Println(err)
		}
		t.Execute(w, nil)
	})
	http.HandleFunc("/serverList", func(w http.ResponseWriter, r *http.Request) {
		if info, err := server_list(); err == nil {
			fmt.Fprintf(w, string(info))
		}
	})
	http.HandleFunc("/apply", func(w http.ResponseWriter, r *http.Request) {
		s, _ := ioutil.ReadAll(r.Body)
		var data map[string]string
		if err := json.Unmarshal(s, &data); err == nil {
			if err = applyServer(data["server"], data["user"], data["pass"], data["time"]); err != nil {
				fmt.Fprintf(w, "{\"result\":\"failed\"}")
			} else {
				fmt.Fprintf(w, "{\"result\":\"success\"}")
			}
		} else {
			log.Println("get body frase failed ")
		}
	})
	http.ListenAndServe(httpserver, nil)
}
