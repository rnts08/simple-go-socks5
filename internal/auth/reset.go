package auth

import (
	"fmt"

	"vn-socks-proxy/internal/config"
)

func ResetAuthDB(cfg *config.Config) error {
	if cfg.AuthMode != config.ModeLocal {
		return fmt.Errorf("reset only works for local auth mode")
	}
	return ResetDB(cfg.AuthDBPath)
}