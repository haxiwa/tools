package tools

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/big"
	"net"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
	"unsafe"
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
func RandomString(len int) string {
	var container string
	var str = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"
	b := bytes.NewBufferString(str)
	length := b.Len()
	bigInt := big.NewInt(int64(length))
	for i := 0; i < len; i++ {
		randomInt, _ := rand.Int(rand.Reader, bigInt)
		container += string(str[randomInt.Int64()])
	}
	return container
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
func DeleteExtraSpace(s string) string {
	//删除字符串中的多余空格，有多个空格时，仅保留一个空格
	s1 := strings.Replace(s, "  ", " ", -1)      //替换tab为空格
	regstr := "\\s{2,}"                          //两个及两个以上空格的正则表达式
	reg, _ := regexp.Compile(regstr)             //编译正则表达式
	s2 := make([]byte, len(s1))                  //定义字符数组切片
	copy(s2, s1)                                 //将字符串复制到切片
	spc_index := reg.FindStringIndex(string(s2)) //在字符串中搜索
	for len(spc_index) > 0 {                     //找到适配项
		s2 = append(s2[:spc_index[0]+1], s2[spc_index[1]:]...) //删除多余空格
		spc_index = reg.FindStringIndex(string(s2))            //继续在字符串中搜索
	}
	return string(s2)
}

//================================================================================
func Post_json(url string, header map[string]string, song map[string]interface{}) string {

	bytesData, err := json.Marshal(song)
	if err != nil {
		fmt.Println(err.Error())
		return "0"
	}
	reader := bytes.NewReader(bytesData)
	request, err := http.NewRequest("POST", url, reader)
	if err != nil {
		fmt.Println(err.Error())
		return "0"
	}
	for k, v := range header {
		request.Header.Set(k, v)
	}
	request.Header.Set("Content-Type", "application/json;charset=UTF-8")
	client := http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		fmt.Println(err.Error())
		return "0"
	}
	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err.Error())
		return "0"
	}
	//byte数组直接转成string，优化内存
	str := (*string)(unsafe.Pointer(&respBytes))
	return *str
}

//================================================================================
//================================================================================
//================================================================================
