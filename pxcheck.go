package main

import (
	"fmt"
	"math/rand"
	"net"
	"os"
	"regexp"
	"time"

	"golang.org/x/net/proxy"
)

// resolved
var dns = map[string][]string{
	"ifconfig.me": []string{
		"153.121.72.211",
		"153.121.72.212",
	},
	"icanhazip.com": []string{
		"45.63.64.111",
		"144.202.71.30",
	},
	"ident.me": []string{
		"176.58.123.25",
	},
	"whatismyip.akamai.com": []string{
		"88.221.216.201",
		"92.123.155.221",
	},
	"wgetip.com": []string{
		"104.31.78.175",
		"104.31.79.175",
	},
	"ip.tyk.nu": []string{
		"144.76.253.225",
	},
	"bot.whatismyipaddress.com": []string{
		"66.171.248.178",
	},
	"eth0.me": []string{
		"99.73.32.195",
	},
	"alma.ch": []string{
		"85.195.225.172",
	},
	"api.infoip.io": []string{
		"130.211.46.130",
	},
	"api.ipify.org": []string{
		"23.23.114.123",
		"50.19.222.19",
		"50.19.229.252",
		"54.235.183.188",
		"54.243.136.64",
		"184.73.209.86",
	},
	"canhazip.com": []string{
		"104.16.77.150",
		"104.16.78.150",
		"104.16.79.150",
		"104.16.80.150",
		"104.16.81.150",
	},
	"checkip.amazonaws.com": []string{
		"34.224.244.1",
		"34.226.30.197",
		"34.232.4.85",
		"54.84.25.24",
	},
	"ipinfo.io": []string{
		"216.239.32.21",
		"216.239.34.21",
		"216.239.36.21",
		"216.239.38.21",
	},
	"smart-ip.net": []string{
		"193.178.146.17",
	},
}

func DNSResolv(host string) (bool, string) {
	val, ok := dns[host]
	if !ok {
		return false, ""
	}
	return true, val[rand.Intn(len(val))]
}

var testers = []string{
	"icanhazip.com",
	"whatismyip.akamai.com",
	"wgetip.com",
	"ip.tyk.nu",
	"bot.whatismyipaddress.com",
	"api.ipify.org",
	"canhazip.com",
	"checkip.amazonaws.com",

	//  ifconfig.me
	//  ident.me
	//  eth0.me
	//  ---
	//+ ipinfo.io/ip
	//  smart-ip.net/myip
	//  alma.ch/myip.cgi
	//+ api.infoip.io/ip
}

//DNS
//dig +short myip.opendns.com @resolver1.opendns.com

type ProxyStats struct {
	ConnectTime time.Duration
	TotalTime   time.Duration
	ExtIP       string
	Tester      string
	Message     string
}

type directTimeout struct{}

// Direct is a direct proxy: one that makes network connections directly.
var DirectTimeout = directTimeout{}

func (directTimeout) Dial(network, addr string) (net.Conn, error) {
	return net.DialTimeout(network, addr, time.Second*5)
}

func CheckProxy(address string) ProxyStats {
	// create a socks5 dialer
	dialer, err := proxy.SOCKS5("tcp", address, nil, DirectTimeout)
	if err != nil {
		return ProxyStats{Message: err.Error()}
	}

	stats := ProxyStats{}
	tester := testers[rand.Intn(len(testers))]
	ok, tester_addr := DNSResolv(tester)
	stats.Tester = tester
	if !ok {
		fmt.Errorf("Error resolving tester. Tester: %s", tester)
		stats.Message = err.Error()
		return stats
	}
	tm_start := time.Now()
	conn, err := dialer.Dial("tcp", tester_addr+":80")
	if err != nil {
		stats.Message = err.Error()
		return stats
	}
	defer conn.Close()

	if err != nil {
		stats.Message = err.Error()
		return stats
	}
	stats.ConnectTime = time.Now().Sub(tm_start)

	message := fmt.Sprintf("GET / HTTP/1.1\r\nHost: %s\r\nUser-Agent: Mozilla/5.0 (Windows NT 6.1; WOW64; rv:23.0) Gecko/20100101 Firefox/23.0\r\n\r\n",
		tester)
	_ = conn.SetDeadline(time.Now().Add(time.Second * 5))
	conn.Write([]byte(message))

	buff := make([]byte, 1024)
	_, err = conn.Read(buff)
	if err != nil {
		stats.Message = err.Error()
		return stats
	}
	stats.TotalTime = time.Now().Sub(tm_start)
	re := regexp.MustCompile(`\r\n\r\n([0-9\.]+)`)
	data := re.FindStringSubmatch(string(buff))
	if len(data) > 1 {
		stats.ExtIP = data[1]
	}

	return stats
}

func main() {
	rand.Seed(time.Now().UnixNano())
	if len(os.Args) < 2 {
		fmt.Printf("Usage: %s <proxy>", os.Args[0])
		return
	}
	stats := CheckProxy(os.Args[1])
	fmt.Printf("%d\t%d\t%d\t%s\t%s\t%s\t%s\n",
		stats.ConnectTime,
		stats.TotalTime,
		stats.TotalTime-stats.ConnectTime,
		os.Args[1], stats.ExtIP, stats.Tester, stats.Message)
}
