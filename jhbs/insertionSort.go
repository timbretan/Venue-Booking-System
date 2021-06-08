package jhbs

func insertionSort(items []interface{},
	sortCriterion func(a interface{}, b interface{}) bool) {
	var n = len(items)
	for i := 1; i < n; i++ {
		j := i
		for j > 0 {
			// if items[j-1] > items[j] {
			if sortCriterion(items[j-1], items[j]) {
				items[j-1], items[j] = items[j], items[j-1]
			}
			j = j - 1
		}
	}
}
