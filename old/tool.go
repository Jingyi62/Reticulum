package main

import (
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"
)

const (
	MaxUint = ^uint(0)
	MinUint = 0
	MaxInt  = int(MaxUint >> 1)
	MinInt  = -(MaxInt - 1)
)

func StructToBSONMap(st interface{}) (m map[string]interface{}) {

	s := reflect.ValueOf(st).Elem()
	typeOfT := s.Type()

	m = make(map[string]interface{})

	for i := 0; i < s.NumField(); i++ {

		field := s.Field(i)
		typeField := typeOfT.Field(i)

		fieldName := strings.Split(typeField.Tag.Get("bson"), ",")[0]

		if fieldName == "" {
			fieldName = typeField.Name
		}

		m[fieldName] = field.Interface()
	}

	return
}

func ArrayOfBytes(i int, b byte) (p []byte) {

	for i != 0 {

		p = append(p, b)
		i--
	}
	return
}

func FitBytesInto(d []byte, i int) []byte {

	if len(d) < i {

		dif := i - len(d)

		return append(ArrayOfBytes(dif, 0), d...)
	}

	return d[:i]
}

func StripByte(d []byte, b byte) []byte {

	for i, bb := range d {

		if bb != b {
			return d[i:]
		}
	}

	return nil
}

func IsNil(v interface{}) bool {
	return reflect.ValueOf(v).IsNil()
}

func DecodeJSON(r io.Reader, t interface{}) (err error) {

	err = json.NewDecoder(r).Decode(t)
	return
}

func SHA1(data []byte) string {

	hash := sha1.New()
	hash.Write(data)
	return SHAString(hash.Sum(nil))
}

func SHA256(data []byte) []byte {
	hash := sha256.New()
	hash.Write(data)
	return hash.Sum(nil)
}

func SHAString(data []byte) string {
	return fmt.Sprintf("%x", data)
}

func RandomString(n int) string {

	alphanum := "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	var bytes = make([]byte, n)
	rand.Read(bytes)
	for i, b := range bytes {
		bytes[i] = alphanum[b%byte(len(alphanum))]
	}
	return string(bytes)
}

func RandomInt(a, b int) int {

	var bytes = make([]byte, 1)
	rand.Read(bytes)

	per := float32(bytes[0]) / 256.0
	dif := Max(a, b) - Min(a, b)

	return Min(a, b) + int(per*float32(dif))
}

func Max(a, b int) int {

	if a >= b {

		return a
	}

	return b
}

func Min(a, b int) int {

	if a <= b {

		return a
	}

	return b
}

func EncodeBase64(data []byte) []byte {

	base64data := []byte{}
	base64.StdEncoding.Encode(base64data, data)
	return base64data
}

func DecodeBase64(base64data []byte) (data []byte) {

	base64.StdEncoding.Decode(data, base64data)
	return
}

func EncodeBigsBase64(is ...*big.Int) []byte {

	arr := []byte{}
	for _, i := range is {
		arr = append(arr, i.Bytes()...)
	}
	return EncodeBase64(arr)
}

func DecodeBigsBase64(d []byte, i int) []*big.Int {

	arr := make([]*big.Int, i)
	is := DecodeBase64(d)
	l := len(is) / i

	for i := range is {

		arr[i] = big.NewInt(0).SetBytes(is[l*i : l*(i+1)])
	}

	return arr
}

func Timeout(i time.Duration) chan bool {

	t := make(chan bool)
	go func() {
		time.Sleep(i)
		t <- true
	}()

	return t
}

func ShuffleBytes(input []byte) []byte {
	output := make([]byte, len(input))
	copy(output, input)

	n := len(output)
	for i := n - 1; i > 0; i-- {
		j, err := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		if err != nil {
			// 处理随机数生成错误
			return input
		}
		index := j.Int64()

		output[i], output[index] = output[index], output[i]
	}

	return output
}
func GenerateRandomNumber() string {
	max := big.NewInt(100)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		// 处理随机数生成错误
		return "1"
	}
	randomNumber := n.Int64() + 1 // 将范围从0-99转换为1-100
	randomString := strconv.Itoa(int(randomNumber))
	return randomString
}
func GetLocalIP() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", fmt.Errorf("failed to get interface addresses: %v", err)
	}

	for _, addr := range addrs {
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				return ipNet.IP.String(), nil
			}
		}
	}

	return "", fmt.Errorf("failed to find local IP")
}

// 获取本地外网IP
func GetPublicIP() (string, error) {
	resp, err := http.Get("https://api.ipify.org?format=text")
	if err != nil {
		return "", fmt.Errorf("failed to get public IP: %v", err)
	}
	defer resp.Body.Close()

	ipStr, err := readResponse(resp)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %v", err)
	}

	ip := strings.TrimSpace(ipStr)

	return ip, nil
}

// 读取HTTP响应内容
func readResponse(resp *http.Response) (string, error) {
	buf := make([]byte, 1024)
	var sb strings.Builder

	for {
		n, err := resp.Body.Read(buf)
		sb.Write(buf[:n])

		if err != nil {
			break
		}
	}

	return sb.String(), nil
}
func IsPrefix(a, b string) bool {
	if len(a) < len(b) {
		return false
	}

	return a[:len(b)] == b
}

func SendAfter(delta int, node *Node, bc *PSchain, cc *CSchain) {
	// 使用time.AfterFunc实现10秒后设置为false的逻辑
	time.AfterFunc(time.Duration(delta)*time.Second, func() {
		node.VotesummaryRound += 1
		node.ifvotesummary = false
		temp_bytes, _ := EncodeVoteSummary(&node.VoteSummary)
		node.VoteSummary.Signature = sha256.Sum256(temp_bytes)
		//将收集到的vote 打包成vote summary进行发送
		cc.VoteSummaryQueue_send <- node.VoteSummary
	})
}
