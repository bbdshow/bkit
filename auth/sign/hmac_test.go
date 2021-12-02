package sign

import (
	"fmt"
	"testing"
)

func TestHmacSha256ToHex(t *testing.T) {
	str := HmacSha256ToHex("accessToken=e84fbed4baba673ff9734b2903378bfd&method=GET&path=/admin/v1/data/textConfig/list1638264941", "61q9zfe8fmmgkdf1ab0cii6gzd2sr5nh")
	fmt.Println(str)
}
