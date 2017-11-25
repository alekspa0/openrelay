package db

import (
	"github.com/notegio/openrelay/types"
	"github.com/jinzhu/gorm"
	"math/big"
	"time"
	"log"
)

const (
	StatusOpen = int64(0)
	StatusFilled = int64(1)
	StatusUnfunded = int64(2)
)

type Order struct {
	types.Order
	CreatedAt time.Time
	UpdatedAt time.Time
	OrderHash []byte `gorm:"primary_key"`
	Status    int64 `gorm:"index"`
}

// Save records the order in the database, defaulting to the specified status.
// Status should either be db.StatusOpen, or db.StatusUnfunded. If the order
// is filled based on order.TakerTokenAmountFilled + order.TakerTokenAmountCancelled
// the status will be recorded as db.StatusFilled regardless of the specified status.
func (order *Order) Save(db *gorm.DB, status int64) (*gorm.DB) {
	order.OrderHash = order.Hash()
	remainingAmount := new(big.Int)
	remainingAmount.SetBytes(order.TakerTokenAmount[:])
	remainingAmount.Sub(remainingAmount, new(big.Int).SetBytes(order.TakerTokenAmountFilled[:]))
	remainingAmount.Sub(remainingAmount, new(big.Int).SetBytes(order.TakerTokenAmountCancelled[:]))
	updates := map[string]interface{}{
		"taker_token_amount_filled": order.TakerTokenAmountFilled,
		"taker_token_amount_cancelled": order.TakerTokenAmountCancelled,
		"status": status,
		"update_at": time.Now(),
	}
	if remainingAmount.Cmp(new(big.Int).SetInt64(0)) <= 0 {
		updates["status"] = StatusFilled
	}
	updateScope := db.Model(Order{}).Where("order_hash = ?", order.OrderHash).Updates(updates)
	if updateScope.Error != nil {
		log.Printf(updateScope.Error.Error())
	}
	if updateScope.RowsAffected > 0 {
		return updateScope
	}
	order.CreatedAt = time.Now()
	order.UpdatedAt = order.CreatedAt
	order.Status = status
	return db.Create(order)
}
