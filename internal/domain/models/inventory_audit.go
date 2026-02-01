package models
import (
    "time"
    "github.com/google/uuid"
)
type InventoryAudit struct {
    ID           uuid.UUID `gorm:"type:uuid;primary_key;" json:"id"`
    InventoryID  uuid.UUID `gorm:"type:uuid;not null;index" json:"inventory_id"`
    ProductID    uuid.UUID `gorm:"type:uuid;not null" json:"product_id"`
    OldStock     int    `gorm:"not null" json:"old_stock"`
    NewStock     int    `gorm:"not null" json:"new_stock"`
    ChangeAmount int    `gorm:"not null" json:"change_amount"`
    
    ActionType   string `gorm:"type:varchar(50);not null" json:"action_type"`
    Note         string `gorm:"type:text" json:"note"`
    CreatedAt    time.Time `json:"created_at"`
}
func (InventoryAudit) TableName() string {
    return "inventory_audit"
}