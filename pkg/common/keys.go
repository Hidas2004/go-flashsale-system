package common

import "fmt"

const (
	PrefixProductStock = "product:%s:stock" //product:uuid:stock
	PrefixOrderLock    = "order:%s:lock"    // order:uuid:lock
	PrefixUserOneTime  = "user:%s:action"   // user:uuid:action (Block spam request)
)

// GetProductStockKey trả về key lưu tồn kho cho Redis
func GetProductStockKey(productID string) string {
	return fmt.Sprintf(PrefixProductStock, productID)
}

// GetOrderLockKey trả về key lock Distributed Lock
func GetOrderLockKey(orderID string) string {
	return fmt.Sprintf(PrefixOrderLock, orderID)
}
