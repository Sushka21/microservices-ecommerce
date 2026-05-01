package entity

import "errors"

type CartItem struct {
	SKU   uint32
	Count uint32
}

var (
	ErrProductNotFound   = errors.New("product not found")
	ErrInsufficientStock = errors.New("not enough stock")
	ErrCartIsEmpty       = errors.New("cart is empty")
	ErrItemNotFound      = errors.New("item not found")
)



