/*
	基本类型切片相关操作
*/

package data

// Contains 判断切片 array 中是否包含元素 ele
func Contains[E comparable](array []E, ele E) bool {
	for _, e := range array {
		if e == ele {
			return true
		}
	}
	return false
}

// Insert 向切片中插入元素,返回新的切片
func Insert[E comparable](array []E, index int, elem ...E) []E {
	// 切片长度
	length := len(array)

	// 避免index out range
	if index >= length {
		index = length
	} else if index <= -length {
		index = 0
	} else if index < 0 {
		// 负索引变正索引
		index += length
	}

	// 缓存索引后的切片
	s := append([]E{}, array[index:]...)
	// 拼接
	array = append(array[:index], elem...)
	return append(array, s...)
}

// Pop 根据索引删除切片中的元素，返回新的切片，和删除的元素
func Pop[E comparable](array []E, index int) ([]E, E) {
	// 切片长度
	length := len(array)

	// 避免index out range
	if index >= length {
		index = length
	} else if index <= -length {
		index = 0
	} else if index < 0 {
		// 负索引变正索引
		index += length
	}

	s := array[index]

	return append(array[:index], array[index+1:]...), s
}

// Equal 简单比较两切片内元素是否相同,并返回第一个不相同索引
//	如果相同返回-2，如果长度不同返回-1
func Equal[E comparable](a, b []E) (bool,int) {
	if len(a) != len(b) {
		return false,-1
	}
	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			return false,i
		}
	}
	return true,-2
}

func IsEmpty(array []string) bool {
	if len(array) == 0 {
		return true
	}
	for _, s := range array {
		if s != "" {
			return false
		}
	}
	return true
}
func RepeatIndex[E comparable](array []E) map[E][]int {

	result := map[E][]int{}

	for i, e := range array {
		if _, ok := result[e]; ok {
			result[e] = append(result[e], i)
		} else {
			result[e] = []int{i}
		}
	}

	for key, value := range result {
		if len(value) == 1 {
			delete(result, key)
		}
	}

	return result
}