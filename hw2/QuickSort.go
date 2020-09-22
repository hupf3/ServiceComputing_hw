package Qsort

// 快速排序递归实现
func QuickSort(arr []int, l, r int) {
	pivot := arr[l] // 选取中间值
	pos := l        // 中间值的位置
	i, j := l, r

	for i <= j {
		for j >= pos && arr[j] >= pivot {
			j--
		}
		if j >= pos {
			arr[pos] = arr[j]
			pos = j
		}

		for i <= pos && arr[i] <= pivot {
			i++
		}
		if i <= pos {
			arr[pos] = arr[i]
			pos = i
		}
	}
	arr[pos] = pivot
	if pos-l > 1 {
		QuickSort(arr, l, pos-1)
	}
	if r-pos > 1 {
		QuickSort(arr, pos+1, r)
	}
}
