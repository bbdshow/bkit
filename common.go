package bkit

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
)

// GetCurrentDir 获取当前目录
func GetCurrentDir() (string, error) {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0])) // absolution filepath.Dir(os.Args[0])
	if err != nil {
		return "", err
	}
	return strings.Replace(dir, "\\", "/", -1), nil // just \ replace to /
}

// FilenameExists 判断文件是否存在
func FilenameExists(filename string) bool {
	_, err := os.Stat(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
		return false
	}
	return true
}

// GetLocalIPv4 获取本地机器IPv4
func GetLocalIPv4() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}
	for _, addr := range addrs {
		// 检查ip地址判断是否回环地址
		ipNet, ok := addr.(*net.IPNet)
		if ok && !ipNet.IP.IsLoopback() {
			// 跳过IPV6
			if ipNet.IP.To4() != nil {
				return ipNet.IP.String(), nil
			}
		}
	}
	return "", nil
}

// GetNodeName 获取节点名称
func GetNodeName() (string, error) {
	podNameSpace := strings.TrimSpace(os.Getenv("POD_NAMESPACE"))
	podName := strings.TrimSpace(os.Getenv("POD_NAME"))
	podIP := strings.TrimSpace(os.Getenv("POD_IP"))
	if podNameSpace != "" && podName != "" && podIP != "" {
		return fmt.Sprintf("%s_%s_%s", podNameSpace, podName, podIP), nil
	}
	// 就取本地机器的IP
	localIP, err := GetLocalIPv4()
	if err != nil {
		return "", fmt.Errorf("GetLocalIPv4 Set NodeName %v", err)
	}
	return localIP, nil
}

// SliceFloat convert []float64 to []T
func SliceFloat[T float32 | float64](s []float64) []T {
	val := make([]T, len(s))
	for i, v := range s {
		val[i] = T(v)
	}
	return val
}
func SliceInt[T int8 | int16 | int | int32](s []int64) []T {
	val := make([]T, len(s))
	for i, v := range s {
		val[i] = T(v)
	}
	return val
}
func SliceUint[T uint8 | uint16 | uint | uint32](s []uint64) []T {
	val := make([]T, len(s))
	for i, v := range s {
		val[i] = T(v)
	}
	return val
}

func IsStructPtr(typ string) bool {
	if !strings.HasPrefix(typ, "*") {
		return false
	}
	return strings.Contains(typ, ".")
}
