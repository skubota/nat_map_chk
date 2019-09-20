package main

import (
	"errors"
	"flag"
	"fmt"
	"gortc.io/stun"
	"log"
	"net"
	"os"
	"strings"
)

var init_port = flag.Int("i", 10000, "initial source port")
var count = flag.Int("c", 3, "count")
var servers = flag.String("s", "stun.l.google.com:19302,stun1.l.google.com:19302,stun2.l.google.com:19302,stun3.l.google.com:19302,stun4.l.google.com:19302", "STUN Servers")
var DEBUG = flag.Bool("D", false, "debug mode")

func main() {

	flag.Parse()

	service := *init_port
	srvs := []string{}
	for _, str := range strings.Split(*servers, ",") {
		srvs = append(srvs, str)
	}

	ip, err := get_my_ipaddress()
	if err != nil {
		fmt.Println(err)
	}
	result := make(map[int]map[int]string)

	for i := 0; i < *count; i++ {

		result[i] = make(map[int]string)
		udpAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", service+i))
		if err != nil {
			log.Fatalln(err)
			os.Exit(1)
		}
		conn, err := net.ListenUDP("udp", udpAddr)
		if err != nil {
			continue
			//log.Fatalln(err)
			//os.Exit(1)
		}

		for j := 0; j < len(srvs); j++ {
			xor := stun_check(conn, srvs[j])
			if *DEBUG {
				log.Printf("Src: %s:%d, Dst: %s, Res: %s:%d\n", ip, service+i, srvs[j], xor.IP, xor.Port)
			}
			result[i][j] = fmt.Sprintf("%s:%d", xor.IP, xor.Port)
		}
	}

	EXTERNAL := ""
	NAT := false
	MAPPING := ""

	for k := range result {
		for l := range result[k] {
			s := strings.Split(result[k][l], ":")

			// TEST1
			if ip == s[0] && fmt.Sprintf("%d", service+k) == s[1] {
				NAT = false
			} else {
				NAT = true
			}
			EXTERNAL = s[0]

			// TEST2
			if len(result[k][l-1]) > 0 {
				m := strings.Split(result[k][l-1], ":")
				if s[1] == m[1] {
					MAPPING = "EIM"
				} else {
					MAPPING = "APDM"
				}
			}
		}
	}

	fmt.Printf("My IP  : %s\n", ip)
	fmt.Printf("Ext IP : %s\n", EXTERNAL)
	fmt.Printf("NAT    : %v\n", NAT)
	if NAT {
		fmt.Printf("   MAPPING : %v\n", MAPPING)
	}

}

func stun_check(conn *net.UDPConn, srv string) stun.XORMappedAddress {
	var xorAddr stun.XORMappedAddress

	message := stun.MustBuild(stun.TransactionID, stun.BindingRequest)
	addr, err := net.ResolveUDPAddr("udp", srv)
	conn.WriteTo(message.Raw, addr)

	buf := make([]byte, 1024)
	n, addr, err := conn.ReadFromUDP(buf)
	if err != nil {
		panic(err)
	}
	message.Length = uint32(n)
	message.Raw = buf
	message.Decode()
	if err := xorAddr.GetFrom(message); err != nil {
		panic(err)
	}
	return xorAddr
	//return fmt.Sprintf("%s:%d", xorAddr.IP, xorAddr.Port)
}

func get_my_ipaddress() (string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue // interface down
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue // loopback interface
		}
		addrs, err := iface.Addrs()
		if err != nil {
			return "", err
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue // not an ipv4 address
			}
			return ip.String(), nil
		}
	}
	return "", errors.New("are you connected to the network?")
}
