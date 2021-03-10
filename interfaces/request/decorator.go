package request

import (
	"github.com/DmitriBeattie/custom-framework/abstract/stack"
	"net/http"
)

type RequestDecorator interface {
	Decorate(r *http.Request) *http.Request
}

type DecoratorChain struct {
	*stack.Stack
}

func New() *DecoratorChain {
	return &DecoratorChain{stack.New()}
}

func (dChain *DecoratorChain) Next(requestDecorator RequestDecorator) *DecoratorChain {
	dChain.Push(requestDecorator)

	return dChain
}

func (dChain *DecoratorChain) Decorate(r *http.Request) *http.Request {
	dec := dChain.Pop()
	if dec == nil {
		return r
	}

	return dChain.Decorate(dec.(RequestDecorator).Decorate(r))
}