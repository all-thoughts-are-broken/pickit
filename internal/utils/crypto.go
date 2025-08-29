package utils

import (
	"crypto/md5"
	"fmt"
)

func getMd5(data string) string {
	result := md5.Sum([]byte(data))
	return fmt.Sprintf("%x", result)
}

/*
GetNum 计算图片被切割的刀数。
参数:

	scrambleId - scrambleId
	aid - 车牌号
	filename - 文件名（不要后缀名）

返回值:

	图片被切割的刀数
*/
func GetNum(scrambleId, aid int, filename string) int {
	if aid < scrambleId {
		return 0
	} else if aid < 268850 {
		return 10
	} else {
		var x int
		if aid < 421926 {
			x = 10
		} else {
			x = 8
		}
		s := fmt.Sprintf("%d%s", aid, filename)
		hash := getMd5(s)
		lastChar := hash[len(hash)-1]
		num := int(lastChar)
		num %= x
		num = num*2 + 2
		return num
	}
}
