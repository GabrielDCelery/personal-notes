package main

import "fmt"

func modifyArray(value [3]int) {
	value[0] = 4
}

func modifySlice(value []int) {
	value[0] = 4
}

func appendToSlice(v []int) {
	v = append(v, 0)
}

func appendToSlicePtr(v *[]int) {
	*v = append(*v, 0)
}

func appendToMap(v map[string]int) {
	v["appended"] = 3
}

func reassingString(s *string) {
	*s = "somethingelse"
}

func main() {
	a := [3]int{0, 1, 2}
	modifyArray(a)
	fmt.Printf("modifyArray - arr passed as value %+v should stay {0,1,2}\n", a)

	s := []int{0, 1, 2, 3}
	modifySlice(s)
	fmt.Printf("modifySlice - slice passed as value %+v shoudld become {4,1,2,3}\n", s)

	s1 := []int{0, 1, 2, 3}
	appendToSlice(s1)
	fmt.Printf("appendToSlice - slice passed as value %+v shoudld stay {0,1,2,3}\n", s1)

	s2 := []int{0, 1, 2, 3}
	appendToSlicePtr(&s2)
	fmt.Printf("appendToSlicePtr - slice passed as pointer %+v shoudld become {0,1,2,3,0}\n", s2)

	m1 := map[string]int{"a": 1}
	appendToMap(m1)
	fmt.Printf("appendToMap - %+v should become map[a:1 appended:3]\n", m1)

	str1 := "foo"
	reassingString(&str1)
	fmt.Printf("reassingString - %s\n", str1)
}
