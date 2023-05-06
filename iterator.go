package main

// Iterator is used in the parser and interpreter
type Iterator[A any] struct {
	index   int
	current A
	array   []A
}

func new_iterator[A any](array []A) Iterator[A] {
	return Iterator[A]{0, array[0], array}
}

func (iter *Iterator[A]) set_current(index int) A {
	iter.index = index
	iter.current = iter.array[index]
	return iter.current
}
func (iter *Iterator[A]) advance() A {
	iter.index++
	iter.current = iter.array[iter.index]
	return iter.current
}
func (iter *Iterator[A]) move_index(i int) {
	iter.index = i
}
func (iter *Iterator[A]) peek() A {
	return iter.array[iter.index+1]
}
func (iter *Iterator[A]) can_peek() bool {
	return iter.index+1 < len(iter.array)
}

// the iterator's index must be on the count_obj at start
func go_to_boundary(iter *Iterator[Token], count_obj, end string) []Token {
	count := 0
	seen_objects := []Token{}
	for {
		// print_comment(iter.current, count)
		if iter.current.value == count_obj {
			count++
		} else if iter.current.value == end {
			count--
		}
		if count == 0 {
			return seen_objects
		}
		seen_objects = append(seen_objects, iter.advance())
	}
}

// the iterator's index must be on the count_obj at start
func search_boundaries(iter Iterator[Token], count_obj string, end string) int {
	count := 0
	for {
		// print_comment(iter.current, count)
		if iter.current.value == count_obj {
			count++
		} else if iter.current.value == end {
			count--
		}
		if count == 0 {
			return iter.index
		}
		iter.advance()
	}
}

// these are currently unused

// func (iter *Iterator[A]) go_back() A {
// 	iter.index--
// 	iter.current = iter.array[iter.index]
// 	return iter.current
// }
func (iter *Iterator[A]) back() A {
	return iter.array[iter.index-1]
}
