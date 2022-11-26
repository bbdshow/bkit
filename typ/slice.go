package typ

import "reflect"

// SliceRemoveDuplicate 切片去重
func SliceRemoveDuplicate(x interface{}, fn func(i, j int) bool) {
	v := reflect.ValueOf(x)
	if v.Kind() != reflect.Ptr {
		return
	}
	if v.Elem().Kind() != reflect.Slice {
		return
	}
	if v.Elem().Len() <= 1 {
		return
	}
	i := 0
	j := 1
	for {
		if i >= v.Elem().Len()-1 {
			break
		}

		isDuplicate := false
		for {
			if j >= v.Elem().Len() {
				break
			}
			if fn(i, j) {
				// 重复
				isDuplicate = true
				break
			}
			j++
		}
		if isDuplicate {
			//通过slice index 去掉重复元素
			//warning: 这里会产生申请内存扩容等操作，所以针对特别大的Slice，可能会有更高效的做法
			v.Elem().Set(reflect.AppendSlice(v.Elem().Slice(0, j), v.Elem().Slice(j+1, v.Elem().Len())))
			continue
		}
		i++
		j = i + 1
	}
}
