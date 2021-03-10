package api

import (
	"delivery-report/framework/abstract/stack"
	"net/http"
)

type MiddlewareFunc func(next http.HandlerFunc) http.HandlerFunc

type MiddlewareFuncElement struct {
	f    MiddlewareFunc
	prev *MiddlewareFuncElement
}

type MiddlewareChain struct {
	*stack.Stack
}

func NewMiddlewareChain() *MiddlewareChain {
	return &MiddlewareChain{
		Stack: stack.New(),
	}
}

func (m *MiddlewareChain) Next(mfw MiddlewareFunc) *MiddlewareChain {
	m.Push(mfw)

	return m
}

func (m *MiddlewareChain) completeChain(next http.HandlerFunc) http.HandlerFunc {
	data := m.Pop()
	if data == nil {
		return next
	}
	f := data.(MiddlewareFunc)

	return m.completeChain(f(next))
}

func (m *MiddlewareChain) Handle(next http.HandlerFunc) http.HandlerFunc {
	return m.completeChain(next)
}
