package haxiwa

import (
	"fmt"
	"net"
)

const banner = `

  ___ ___               .__                 
 /   |   \_____  ___  __|__|_  _  _______   
/    ~    \__  \ \  \/  /  \ \/ \/ /\__  \  
\    Y    // __ \_>    <|  |\     /  / __ \_
 \___|_  /(____  /__/\_ \__| \/\_/  (____  /
       \/      \/      \/                \/ 

`

//================================================================================
// show cmd banner
func ShowBanner() {
	fmt.Println(banner)
}

//================================================================================
//CIDR to ip range
func inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}
func Hosts(cidr string) ([]string, error) {
	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, err
	}

	var ips []string
	for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); inc(ip) {
		ips = append(ips, ip.String())
	}
	// remove network address and broadcast address
	return ips[1 : len(ips)-1], nil
}

//================================================================================
//================================================================================
//================================================================================
//================================================================================
//================================================================================
//================================================================================
//================================================================================
//================================================================================
//================================================================================
//================================================================================
//================================================================================
//================================================================================
//================================================================================
//================================================================================
func main() {
	ShowBanner()
}
