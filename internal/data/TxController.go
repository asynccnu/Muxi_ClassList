package data

import (
	"class/internal/biz"
	"context"

	"gorm.io/gorm"
)

type TxController struct {
}
func NewTxController() biz.TxController{
	return &TxController{}
}
func (t *TxController) Begin(ctx context.Context, db *gorm.DB) *gorm.DB {
	return db.Begin()
}

func (t *TxController) RollBack(ctx context.Context, tx *gorm.DB) {
	tx.Rollback()
}

func (t *TxController) Commit(ctx context.Context, tx *gorm.DB) error {
	return tx.Commit().Error
}
