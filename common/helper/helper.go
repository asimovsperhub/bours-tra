package helper

import (
	"crypto/md5"
	"database/sql/driver"
	"encoding/hex"
	"fmt"
	"io"
	"math/big"
	"net"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// LocalTime 时间字段格式化输出
type LocalTime time.Time

// BirthdayTime 时间字段格式化输出
type BirthdayTime time.Time

// MarshalJSON LocalTime格式化
func (t *LocalTime) MarshalJSON() ([]byte, error) {
	tTime := time.Time(*t)
	return []byte(fmt.Sprintf("\"%v\"", tTime.Format("2006-01-02 15:04:05"))), nil
}

func (t *LocalTime) Scan(v interface{}) error {
	if value, ok := v.(time.Time); ok {
		*t = LocalTime(value)
		return nil
	}
	return fmt.Errorf("can not convert %v to timestamp", v)
}

func (t LocalTime) Value() (driver.Value, error) {
	var zeroTime time.Time
	tlt := time.Time(t)
	//判断给定时间是否和默认零时间的时间戳相同
	if tlt.UnixNano() == zeroTime.UnixNano() {
		return nil, nil
	}
	return tlt, nil
}

// MarshalJSON BirthdayTime格式化
func (t *BirthdayTime) MarshalJSON() ([]byte, error) {
	tTime := time.Time(*t)
	return []byte(fmt.Sprintf("\"%v\"", tTime.Format("2006-01-02"))), nil
}

func (t *BirthdayTime) Scan(v interface{}) error {
	if value, ok := v.(time.Time); ok {
		*t = BirthdayTime(value)
		return nil
	}
	return fmt.Errorf("can not convert %v to timestamp", v)
}

func (t BirthdayTime) Value() (driver.Value, error) {
	var zeroTime time.Time
	tlt := time.Time(t)
	//判断给定时间是否和默认零时间的时间戳相同
	if tlt.UnixNano() == zeroTime.UnixNano() {
		return nil, nil
	}
	return tlt, nil
}

func ParseBank(bank string) string {
	let := len(bank) - 4
	reg := "\\d{" + strconv.Itoa(let) + "}(\\d{4})"
	re := regexp.MustCompile(reg)
	var str string
	for i := 0; i < len(bank)-4; i++ {
		str += "*"
	}
	str = str + "$1"
	return re.ReplaceAllStringFunc(bank, func(m string) string {
		return re.ReplaceAllString(m, str)
	})
}

// Md5 MD5加密
func Md5(str string, salt ...interface{}) (CryptStr string) {
	//判断salt盐是否存在
	if l := len(salt); l > 0 {
		slice := make([]string, l+1)
		str = fmt.Sprintf(str+strings.Join(slice, "%v"), salt...)
	}
	return fmt.Sprintf("%x", md5.Sum([]byte(str)))
}

func GeneratorMD5(code string) string {
	MD5 := md5.New()
	_, _ = io.WriteString(MD5, code)
	return hex.EncodeToString(MD5.Sum(nil))
}

// InetAtoi 字符串IP转int64
func InetAtoi(ip string) int64 {
	ret := big.NewInt(0)
	ret.SetBytes(net.ParseIP(ip).To4())
	return ret.Int64()
}

// InetItoa int64的IP转字符串IP
func InetItoa(ip int64) string {
	return fmt.Sprintf("%d.%d.%d.%d", byte(ip>>24), byte(ip>>16), byte(ip>>8), byte(ip))
}
