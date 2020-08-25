package main

import (
	"encoding/base64"
	"encoding/binary"
	"log"
	"net"
	"syscall"
	"unsafe"
)

type Winsize struct {
	Height uint16
	Width  uint16
	x      uint16
	y      uint16
}

func parseDims(b []byte) (uint32, uint32) {
	w := binary.BigEndian.Uint32(b)
	h := binary.BigEndian.Uint32(b[4:])
	return w, h
}

func SetWinsize(fd uintptr, w, h uint32) {
	ws := &Winsize{Width: uint16(w), Height: uint16(h)}
	syscall.Syscall(syscall.SYS_IOCTL, fd, uintptr(syscall.TIOCSWINSZ), uintptr(unsafe.Pointer(ws)))
}
func get_local_ip() []byte {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		log.Println(err)
		return nil
	}
	for _, address := range addrs {
		// check loopback address
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				log.Println(ipnet.IP.String())
				return []byte(ipnet.IP.String())
			}

		}
	}
	return nil
}
func private_encode(str string) string {
	magic_str := []byte("z_tstaflafl)aldfjaljl*afldfjalj")
	str_slice := []byte(str)
	for i := 0; i < len(str_slice); i++ {
		str_slice[i] = str_slice[i] + magic_str[i%len(magic_str)]
	}
	return base64.StdEncoding.EncodeToString(str_slice)
}
func private_decode(str string) string {
	magic_str := []byte("z_tstaflafl)aldfjaljl*afldfjalj")
	origin_str, err := base64.StdEncoding.DecodeString(str)
	if err == nil {
		for i := 0; i < len(origin_str); i++ {
			origin_str[i] = origin_str[i] - magic_str[i%len(magic_str)]
		}
	}
	return string(origin_str)
}
