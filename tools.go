package tools

import (
	"bufio"
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/StackExchange/wmi"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"golang.org/x/sys/windows"
	"io"
	"io/ioutil"
	"log"
	"math/big"
	"net"
	"net/http"
	"os"
	"regexp"
	"sort"
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
func SetCookies(url, domain string, cookies []string) chromedp.Tasks {
	//创建一个chrome任务
	//targetNameJS := `document.querySelector("#transferContainer > div.input_list.bor_b_s.m_t29 > ul > li:nth-child(1) > div > div.c > div.drop-down-list.por.sel_list2.add.fl > div > input")`
	//targetNoJS := `document.querySelector("#payeeBankCardInput")`
	//transferMoneyJS := `document.querySelector("#transAmt")`
	//nextStepJS := `document.querySelector("#transferContainer > div.input_list.bor_b_s.m_t29 > ul > li:nth-child(10) > div > a")`
	return chromedp.Tasks{
		//ActionFunc是一个适配器，允许使用普通函数作为操作。
		chromedp.ActionFunc(func(ctx context.Context) error {
			// 设置Cookie存活时间
			//expr := cdp.TimeSinceEpoch(time.Now().Add(180 * 24 * time.Hour))
			// 添加Cookie到chrome
			for i := 0; i < len(cookies); i += 2 {
				//SetCookie使用给定的cookie数据设置一个cookie； 如果存在，可能会覆盖等效的cookie。
				success, err := network.SetCookie(cookies[i], cookies[i+1]).
					// 设置cookie到期时间
					//WithExpires(&expr).
					// 设置cookie作用的站点
					WithDomain(domain). //访问网站主体
					//// 设置httponly,防止XSS攻击
					//WithHTTPOnly(true).
					////Do根据提供的上下文执行Network.setCookie。
					Do(ctx)
				if err != nil {
					log.Println(err)
					return err
				}
				if !success {
					return fmt.Errorf("could not set cookie %q to %q", cookies[i], cookies[i+1])
				}
			}
			return nil
		}),
		// 跳转指定的url地址
		chromedp.Navigate(url),
		chromedp.Sleep(time.Second * 5),
		//chromedp.WaitVisible(targetNameJS, chromedp.ByJSPath),
		//chromedp.SendKeys(targetNameJS, "李江", chromedp.ByJSPath),
		//chromedp.SendKeys(targetNoJS, "6227000010170085031", chromedp.ByJSPath),
		//chromedp.SendKeys(transferMoneyJS, "0.01", chromedp.ByJSPath),
		//chromedp.Sleep(time.Second * 1),
		//chromedp.Click(nextStepJS, chromedp.ByJSPath),
	}
}

//================================================================================
func AgreeDisclaimer() {
	ShowBanner()
	var flag int
	fmt.Println(`免责声明：`)
	fmt.Println(`       本软件仅供个人测试或帮助残障人士代替手动操作，请勿用于任何非法用途，产生任何法律纠纷与作者无关！继续使用即代表同意本协议。`)
	fmt.Println(`请输入：1（同意） 2（拒绝）`)
	fmt.Scanln(&flag)
	if flag != 1 {
		os.Exit(0)
	}

}

//================================================================================
func Unicode2utf8(source string) string {
	var res = []string{""}
	sUnicode := strings.Split(source, "\\u")
	var context = ""
	for _, v := range sUnicode {
		var additional = ""
		if len(v) < 1 {
			continue
		}
		if len(v) > 4 {
			rs := []rune(v)
			v = string(rs[:4])
			additional = string(rs[4:])
		}
		temp, err := strconv.ParseInt(v, 16, 32)
		if err != nil {
			context += v
		}
		context += fmt.Sprintf("%c", temp)
		context += additional
	}
	res = append(res, context)
	urlstr := strings.Join(res, "")
	return urlstr
}

//================================================================================
func Unicode2Gbk(ustr string) string {
	sUnicodev := strings.Split(ustr, "\\u")
	var context string
	for _, v := range sUnicodev {
		if len(v) < 1 {
			continue
		}
		temp, err := strconv.ParseInt(v, 16, 32)
		if err != nil {
			panic(err)
		}
		context += fmt.Sprintf("%c", temp)
	}
	return context
}

//================================================================================
func GetHmacSha256(message, secretKey string) string {
	key := []byte(secretKey)
	h := hmac.New(sha256.New, key)
	h.Write([]byte(message))
	//	fmt.Println(h.Sum(nil))
	sha := h.Sum(nil)
	return base64.StdEncoding.EncodeToString([]byte(sha))
}

//================================================================================
func CheckMac(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	if (a == nil) != (b == nil) {
		return false
	}

	for key, value := range a {
		if value != b[key] {
			return false
		}
	}

	return true
}

//================================================================================
func GetMacAddrs() (macAddrs []string) {
	netInterfaces, err := net.Interfaces()
	if err != nil {
		fmt.Printf("fail to get net interfaces: %v", err)
		return macAddrs
	}

	for _, netInterface := range netInterfaces {
		macAddr := netInterface.HardwareAddr.String()
		if len(macAddr) == 0 {
			continue
		}

		macAddrs = append(macAddrs, macAddr)
	}
	return macAddrs
}

//================================================================================
func GetPhysicalID() string {
	var ids []string
	if guid, err := getMachineGuid(); err != nil {
		panic(err.Error())
	} else {
		ids = append(ids, guid)
	}
	if cpuinfo, err := getCPUInfo(); err != nil && len(cpuinfo) > 0 {
		panic(err.Error())
	} else {
		ids = append(ids, cpuinfo[0].VendorID+cpuinfo[0].PhysicalID)
	}
	if mac, err := getMACAddress(); err != nil {
		panic(err.Error())
	} else {
		ids = append(ids, mac)
	}
	sort.Strings(ids)
	idsstr := strings.Join(ids, "|/|")
	return GetMd5String(idsstr, true, true)
}

func getMACAddress() (string, error) {
	netInterfaces, err := net.Interfaces()
	if err != nil {
		panic(err.Error())
	}
	mac, macerr := "", errors.New("无法获取到正确的MAC地址")
	for i := 0; i < len(netInterfaces); i++ {
		//fmt.Println(netInterfaces[i])
		if (netInterfaces[i].Flags&net.FlagUp) != 0 && (netInterfaces[i].Flags&net.FlagLoopback) == 0 {
			addrs, _ := netInterfaces[i].Addrs()
			for _, address := range addrs {
				ipnet, ok := address.(*net.IPNet)
				//fmt.Println(ipnet.IP)
				if ok && ipnet.IP.IsGlobalUnicast() {
					// 如果IP是全局单拨地址，则返回MAC地址
					mac = netInterfaces[i].HardwareAddr.String()
					return mac, nil
				}
			}
		}
	}
	return mac, macerr
}

type cpuInfo struct {
	CPU        int32  `json:"cpu"`
	VendorID   string `json:"vendorId"`
	PhysicalID string `json:"physicalId"`
}

type win32_Processor struct {
	Manufacturer string
	ProcessorID  *string
}

func getCPUInfo() ([]cpuInfo, error) {
	var ret []cpuInfo
	var dst []win32_Processor
	q := wmi.CreateQuery(&dst, "")
	//fmt.Println(q)
	if err := wmiQuery(q, &dst); err != nil {
		return ret, err
	}

	var procID string
	for i, l := range dst {
		procID = ""
		if l.ProcessorID != nil {
			procID = *l.ProcessorID
		}

		cpu := cpuInfo{
			CPU:        int32(i),
			VendorID:   l.Manufacturer,
			PhysicalID: procID,
		}
		ret = append(ret, cpu)
	}

	return ret, nil
}

// WMIQueryWithContext - wraps wmi.Query with a timed-out context to avoid hanging
func wmiQuery(query string, dst interface{}, connectServerArgs ...interface{}) error {
	ctx := context.Background()
	if _, ok := ctx.Deadline(); !ok {
		ctxTimeout, cancel := context.WithTimeout(ctx, 3000000000) //超时时间3s
		defer cancel()
		ctx = ctxTimeout
	}

	errChan := make(chan error, 1)
	go func() {
		errChan <- wmi.Query(query, dst, connectServerArgs...)
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-errChan:
		return err
	}
}

func getMachineGuid() (string, error) {
	// there has been reports of issues on 32bit using golang.org/x/sys/windows/registry, see https://github.com/shirou/gopsutil/pull/312#issuecomment-277422612
	// for rationale of using windows.RegOpenKeyEx/RegQueryValueEx instead of registry.OpenKey/GetStringValue
	var h windows.Handle
	err := windows.RegOpenKeyEx(windows.HKEY_LOCAL_MACHINE, windows.StringToUTF16Ptr(`SOFTWARE\Microsoft\Cryptography`), 0, windows.KEY_READ|windows.KEY_WOW64_64KEY, &h)
	if err != nil {
		return "", err
	}
	defer windows.RegCloseKey(h)

	const windowsRegBufLen = 74 // len(`{`) + len(`abcdefgh-1234-456789012-123345456671` * 2) + len(`}`) // 2 == bytes/UTF16
	const uuidLen = 36

	var regBuf [windowsRegBufLen]uint16
	bufLen := uint32(windowsRegBufLen)
	var valType uint32
	err = windows.RegQueryValueEx(h, windows.StringToUTF16Ptr(`MachineGuid`), nil, &valType, (*byte)(unsafe.Pointer(&regBuf[0])), &bufLen)
	if err != nil {
		return "", err
	}

	hostID := windows.UTF16ToString(regBuf[:])
	hostIDLen := len(hostID)
	if hostIDLen != uuidLen {
		return "", fmt.Errorf("HostID incorrect: %q\n", hostID)
	}

	return hostID, nil
}

//生成32位md5字串
func GetMd5String(s string, upper bool, half bool) string {
	h := md5.New()
	h.Write([]byte(s))
	result := hex.EncodeToString(h.Sum(nil))
	if upper == true {
		result = strings.ToUpper(result)
	}
	if half == true {
		result = result[8:24]
	}
	return result
}

//利用随机数生成Guid字串
func UniqueId() string {
	b := make([]byte, 48)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return GetMd5String(base64.URLEncoding.EncodeToString(b), true, false)
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
