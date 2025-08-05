package util

import "net"

// GetLocalIP 获取本机非回环IPv4地址
// 返回:
//
//	string - 第一个找到的非回环IPv4地址
//	error - 获取网络接口地址时发生的错误
func GetLocalIP() (string, error) {
	// 获取本机所有网络接口地址
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}

	// 遍历所有网络接口地址
	for _, address := range addrs {
		// 检查地址是否是IPNet类型且不是回环地址
		// 如果没有找到符合条件的地址，返回空字符串
		// 返回找到的第一个符合条件的IPv4地址
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String(), nil
			}
		}
	}
	return "", nil
}
