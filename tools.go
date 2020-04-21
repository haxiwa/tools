package tools

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/csv"
	"fmt"
	"io"
	"math/rand"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

const banner = `

  ___ ___               .__                 
 /   |   \_____  ___  __|__|_  _  _______   
/    ~    \__  \ \  \/  /  \ \/ \/ /\__  \  
\        // __ \_>    <|  |\     /  / __ \_
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
//router
type Pool struct {
	Queue chan int
	Wg    *sync.WaitGroup
}

func NewPool(cap, total int) *Pool {
	if cap < 1 {
		cap = 1
	}
	p := &Pool{
		Queue: make(chan int, cap),
		Wg:    new(sync.WaitGroup),
	}
	p.Wg.Add(total)
	return p
}

func (p *Pool) AddOne() {
	p.Queue <- 1
}

func (p *Pool) DelOne() {
	<-p.Queue
	p.Wg.Done()
}

//================================================================================
//read csv via colums(int)
func Read_csv(path string, columns int) []string {
	file, err := os.Open(path)
	if err != nil {
		fmt.Println("Error:", err)
		return nil
	}
	defer file.Close()
	reader := csv.NewReader(file)
	var list []string
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			fmt.Println("Error:", err)
			return nil
		}
		list = append(list, record[columns])
	}
	return list
}

//================================================================================
//save string to txtfile
func Save_txt_append(content, filepath string) {
	f, err := os.OpenFile(filepath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	defer f.Close()
	if err != nil {
		// 创建文件失败处理

	} else {
		_, err = f.Write([]byte(content))
		if err != nil {
			// 写入失败处理

		}

	}
}
func Save_txt_cover(content, filepath string) {
	f, err := os.OpenFile(filepath, os.O_WRONLY|os.O_TRUNC, 0600)
	defer f.Close()
	if err != nil {
		fmt.Println(err.Error())
	} else {
		_, err = f.Write([]byte(content))
	}
}

//================================================================================
func Randomstring(n int) string {
	var letters = []byte("qwrtyuioplkjhgfdsazxcvbnmQWERTYUIOPLKJHGFDSAZXCVBNM")
	result := make([]byte, n)
	rand.Seed(time.Now().Unix())
	for i := range result {
		result[i] = letters[rand.Intn(n)]
	}
	return string(result)
}

//================================================================================
//Is a digit string?
func IsSingleDigit(data string) bool {
	digit := []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9"}
	for _, item := range digit {
		if data == item {
			return true
		}
	}
	return false
}

func IsDigit(data string) bool {
	for _, item := range data {
		if IsSingleDigit(string(item)) {
			continue
		} else {
			return false
		}
	}
	return true
}

//================================================================================
//Get string between a and b in c.
func Between(str, starting, ending string) string {
	s := strings.Index(str, starting)
	if s < 0 {
		return ""
	}
	s += len(starting)
	e := strings.Index(str[s:], ending)
	if e < 0 {
		return ""
	}
	return str[s : s+e]
}

//================================================================================
func Readtxt(filepath string) []string {
	f, err := os.Open(filepath)
	if err != nil {
		fmt.Println("os Open error: ", err)
		return nil
	}
	defer f.Close()

	br := bufio.NewReader(f)
	var strlist []string
	for {
		line, _, err := br.ReadLine()
		if err == io.EOF {
			break
		} else if err != nil {
			fmt.Println("br ReadLine error: ", err)
			return nil
		}
		strlist = append(strlist, string(line))

	}
	return strlist
}

//================================================================================
//ip到数字
func ip2Long(ip string) uint32 {
	var long uint32
	binary.Read(bytes.NewBuffer(net.ParseIP(ip).To4()), binary.BigEndian, &long)
	return long
}

//数字到IP
func backtoIP4(ipInt int64) string {
	// need to do two bit shifting and “0xff” masking
	b0 := strconv.FormatInt((ipInt>>24)&0xff, 10)
	b1 := strconv.FormatInt((ipInt>>16)&0xff, 10)
	b2 := strconv.FormatInt((ipInt>>8)&0xff, 10)
	b3 := strconv.FormatInt((ipInt & 0xff), 10)
	return b0 + "." + b1 + "." + b2 + "." + b3
}
func Ip2list(addr1, addr2 string) []string {
	var iplist []string
	ip1 := ip2Long(addr1)
	ip2 := ip2Long(addr2)
	for i := ip1; i <= ip2; i++ {
		i := int64(i)
		iplist = append(iplist, backtoIP4(i))
	}
	return iplist
}

//================================================================================
func WaitTimeout(wg *sync.WaitGroup, timeout time.Duration) bool {
	c := make(chan struct{})
	go func() {
		defer close(c)
		wg.Wait()
	}()
	select {
	case <-c:
		return false // completed normally
	case <-time.After(timeout):
		return true // timed out
	}
}

//================================================================================
//================================================================================
//================================================================================
//================================================================================
//================================================================================
