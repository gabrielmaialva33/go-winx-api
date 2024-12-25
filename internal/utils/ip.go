package utils

import (
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
)

func GetIP(public bool) (string, error) {
	var ip string
	var err error
	if public {
		ip, err = GetPublicIP()
	} else {
		ip, err = GetInternalIP()
	}
	if ip == "" {
		ip = "localhost"
	}
	if err != nil {
		return "localhost", err
	}
	return ip, nil
}

func GetInternalIP() (string, error) {
	conn, err := net.Dial("udp", "1.1.1.1:80")
	if err != nil {
		return "", errors.New("no internet connection")
	}
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			_ = fmt.Errorf("error while closing connection: %v", err)
			return
		}
	}(conn)
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String(), nil
}

func GetPublicIP() (string, error) {
	resp, err := http.Get("https://api.ipify.org?format=text")
	if err != nil {
		return "", err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			_ = fmt.Errorf("error while closing body: %v", err)
			return
		}
	}(resp.Body)
	ip, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if !checkIfIpAccessible(string(ip)) {
		return string(ip), errors.New("PORT is blocked by firewall")
	}
	return string(ip), nil
}

func checkIfIpAccessible(ip string) bool {
	conn, err := net.Dial("tcp", ip+":80")
	if err != nil {
		return false
	}
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			_ = fmt.Errorf("error while closing connection: %v", err)
			return
		}
	}(conn)
	return true
}
