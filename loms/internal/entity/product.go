package entity

import "errors"

var (
	ErrProductNotFound   = errors.New("product not found")
	ErrInsufficientStock = errors.New("insufficient stock")
)

type ProductInfo struct {
	Sku   uint32
	Name  string
	Price uint32
	Count uint64
}
