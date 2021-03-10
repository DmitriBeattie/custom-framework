package stack

import "sync"

type StackElement struct {
	data interface{}
	prev *StackElement
}

type Stack struct {
	top *StackElement
	sync.Mutex
}

func New() *Stack {
	return &Stack{}
}

func (s *Stack) Push(data interface{}) {
	s.Lock()

	elem := &StackElement{
		data: data,
		prev: s.top,
	}

	s.top = elem

	s.Unlock()
}

func (s *Stack) Pop() interface{} {
	s.Lock()
	defer s.Unlock()

	if s.top == nil {
		return nil
	}

	res := s.top

	s.top = res.prev

	return res.data
}
