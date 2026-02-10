package postgres

import (
	"context"

	"gorm.io/gorm"
)

func GetDB(ctx context.Context, originalDB *gorm.DB) *gorm.DB {
	//1 kiểm tra xem trong ví (ctx) có tấm vé VIP("DB_TX") không
	//dòng này làm 2 nhiệm vụ
	//-lấy giá trị ra the0 key "DB_TX"
	//Ép kiểu (Type Assertion) nó về dạng *gorm.DB
	tx, ok := ctx.Value("DB_TX").(*gorm.DB)
	//2 trường hợp có transaction
	if ok {
		return tx
	}
	//3 trường hợp không có
	//nghĩa đây là 1 lệnh đơn lẽ bình thường (vd lấy ds san pham xem chơi)
	//WithContext(ctx) giúp gorm biết khi nào user hủy request thì hủy luôn câu SQ
	return originalDB.WithContext(ctx)

}
