package Qsort

import "testing"

const MAXN = 10

var arr = []int{7, 4, 8, 5, 3, 6, 9, 1, 10, 2}

func TestQuickSort(t *testing.T) {
	QuickSort(arr, 0, MAXN-1)
	expected := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	flag := true
	for i := 0; i < MAXN; i++ {
		if arr[i] != expected[i] {
			flag = false
		}
	}
	if flag == false {
		t.Errorf("expected ")
		for i := 0; i < MAXN; i++ {
			t.Errorf("%d\t", expected[i])
		}
		t.Errorf("but got")
		for i := 0; i < MAXN; i++ {
			t.Errorf("%d\t", arr[i])
		}
	}

}
