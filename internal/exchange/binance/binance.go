package binance

import "log/slog"

type BinanceConfig struct {
	Net string
	Log *slog.Logger
}
