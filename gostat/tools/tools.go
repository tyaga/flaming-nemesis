package tools

import (
	"crypto/md5"
	"fmt"
	"io"
	"sort"
)
///
type Hash map[string]string
func (self *Hash) SortedKeys() []string {
	arr := []string{}
	for k, _ := range *self {
		arr = append(arr, k)
	}
	sort.Strings(arr)
	return arr
}
///
func SignMd5(str string) string {
	h := md5.New()
	io.WriteString(h, str)
	return fmt.Sprintf("%x", h.Sum(nil))
}
///
func PrintOnPanic() {
	if r := recover(); r != nil {
		fmt.Println(r)
	}
}
