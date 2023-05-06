package main

import "fmt"

// replace [start;end] part of array by val
func replace_array_part[A any](arr []A, start, end int, val A) []A {
	new_arr := arr[:start]
	new_arr = append(new_arr, val)
	new_arr = append(new_arr, arr[end+1:]...)
	return new_arr
}
func replace_array_part_by_array[A any](arr1 []A, start, end int, arr2 []A) []A {
	new_arr := arr1[:start]
	new_arr = append(new_arr, arr2...)
	new_arr = append(new_arr, arr1[end:]...)
	return new_arr
}

type Stack[A any] struct {
	array []A
	len   int
}

func new_stack[A any](arr []A) Stack[A] {
	if arr == nil {
		s := Stack[A]{}
		s.len = 0
		return s
	}
	return Stack[A]{arr, len(arr)}
}

func (s *Stack[A]) push(n A) {
	s.array = append(s.array, n)
	s.len++
}
func (s *Stack[A]) pop() A {
	if len(s.array)-1 < 0 {
		panic("can't pop from stack with length 0" + fmt.Sprintf("%v", s))
	}
	ret := s.array[s.len-1]
	s.array = s.array[:s.len-1]
	s.len--
	return ret
}
func (s *Stack[A]) last() A {
	return s.array[len(s.array)-1]
}

// currently unused

// func (s *Stack[A]) peek(n int) A {
// 	return s.array[len(s.array)-n]
// }
