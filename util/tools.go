package util

// 这个文件存放公用的工具方法

// 某个数组是否包含某个字符串
//同PHP in_array
func InArray(arr []string, str string) bool {
	for _, val := range arr {
		if val == str {
			return true
		}
	}
	return false
}

