package port

import "errors"

var (
	ErrProductNotFound    = errors.New("product not found")
	ErrInvalidArgument    = errors.New("invalid argument")
	ErrOrderCreateFailed  = errors.New("order create failed")
	ErrGetStocksFailed    = errors.New("get stoks failed")
	ErrGetProductFailed   = errors.New("get roduct failed")
	ErrListProductsFailed = errors.New("list product failed")
	ErrInsufficientStock  = errors.New("insufficient stock")
)

type ProductInfo struct {
	Name  string
	Price uint32
}
