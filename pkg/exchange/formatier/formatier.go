package formatier

import "strings"

func InstrumentFormatier(exhangeName, instrument string) string {
	switch exhangeName {
	case "okx":
		return strings.ReplaceAll(instrument, "/", "-")
	case "bybit":
		return strings.ToUpper(strings.ReplaceAll(instrument, "/", ""))
	case "binance":
		return strings.ToUpper(strings.ReplaceAll(instrument, "/", ""))
	default:
		return ""
	}
}
