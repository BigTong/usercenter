package usercenter

import (
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

func GetUrlPathArg(path string, index int) string {
	segs := strings.Split(path, "/")
	if index < len(segs) && index >= 0 {
		return segs[index]
	}
	return ""
}

func StringToInt64(src string) int64 {
	ret, err := strconv.ParseInt(src, 10, 64)
	if err != nil {
		panic("convert to int64 get error")
		return -1
	}
	return ret
}

func ReadHttpRequestBody(r *http.Request) ([]byte, error) {
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	defer r.Body.Close()
	return data, nil
}
