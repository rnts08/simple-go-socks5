package accounting

import (
	"fmt"

	"vn-socks-proxy/internal/config"
)

func ResetAccountingDB(cfg *config.Config) error {
	if cfg.AccountingMode != config.ModeLocal && cfg.AccountingMode != config.ModeBoth {
		return fmt.Errorf("reset only works for local or both accounting mode")
	}
	return ResetTrafficDB(cfg.AccountingDBPath)
}