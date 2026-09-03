package servicehttp

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/Flikest/PingVi_backend/internal/delivery/dto"
	"github.com/Flikest/PingVi_backend/pkg/exchange"
	"github.com/gin-gonic/gin"
)

func (s *ServiceExchange) Searchinstruments(ctx *gin.Context) {
	instrument := ctx.Param("instrument")
	if instrument == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "instrument is required"})
		return
	}

	before, _, _ := strings.Cut(instrument, "/")

	okxSymbol := strings.ReplaceAll(instrument, "/", "-")
	bybitSymbol := strings.ReplaceAll(instrument, "/", "")
	binanceSymbol := strings.ReplaceAll(instrument, "/", "")

	var wg sync.WaitGroup
	var mu sync.Mutex
	var result []dto.SearchInstrumentResponse
	var errs []error

	wg.Add(3)

	go func() {
		defer wg.Done()

		okxInstruments, err := exchange.GetOKXInstruments(okxSymbol)
		if err != nil {
			mu.Lock()
			errs = append(errs, fmt.Errorf("OKX instruments: %w", err))
			mu.Unlock()
			return
		}

		ohlc, err := exchange.GetOKXCandles(okxSymbol, "", "")
		if err != nil {
			s.Log.Error("error with fetching okx candles: ", "error", err)
			mu.Lock()
			errs = append(errs, fmt.Errorf("OKX candles: %w", err))
			mu.Unlock()
			return
		}

		if len(okxInstruments.Data) == 0 || len(ohlc) == 0 {
			mu.Lock()
			errs = append(errs, fmt.Errorf("OKX: no data available"))
			mu.Unlock()
			return
		}

		mu.Lock()
		result = append(result, dto.SearchInstrumentResponse{
			ExchangeName:   "OKX",
			Symbol:         okxInstruments.Data[0].InstId,
			LogoURL:        fmt.Sprintf("https://img.logo.dev/crypto/%s?token=%s", before, os.Getenv("DEV_LOGO_PK")),
			OHLCLastCandle: ohlc[len(ohlc)-1],
		})
		mu.Unlock()
	}()

	go func() {
		defer wg.Done()

		bybitInstruments, err := exchange.GetByBitInstruments(bybitSymbol)
		if err != nil {
			mu.Lock()
			errs = append(errs, fmt.Errorf("ByBit instruments: %w", err))
			mu.Unlock()
			return
		}

		ohlc, err := exchange.GeByBitCandles(bybitSymbol, "", "")
		if err != nil {
			s.Log.Error("error with fetching bybit candles: ", "error", err)
			mu.Lock()
			errs = append(errs, fmt.Errorf("ByBit candles: %w", err))
			mu.Unlock()
			return
		}

		if len(bybitInstruments.Symbols) == 0 || len(ohlc) == 0 {
			mu.Lock()
			errs = append(errs, fmt.Errorf("ByBit: no data available"))
			mu.Unlock()
			return
		}

		mu.Lock()
		result = append(result, dto.SearchInstrumentResponse{
			ExchangeName:   "ByBit",
			Symbol:         bybitInstruments.Symbols[0].Symbol,
			LogoURL:        fmt.Sprintf("https://img.logo.dev/crypto/%s?token=%s", before, os.Getenv("DEV_LOGO_PK")),
			OHLCLastCandle: ohlc[len(ohlc)-1],
		})
		mu.Unlock()
	}()

	go func() {
		defer wg.Done()

		binanceInstruments, err := exchange.GetBinanceInstruments(binanceSymbol)
		if err != nil {
			mu.Lock()
			errs = append(errs, fmt.Errorf("Binance instruments: %w", err))
			mu.Unlock()
			return
		}

		ohlc, err := exchange.GetBinanceCandles(binanceSymbol, "", "")
		if err != nil {
			s.Log.Error("error with fetching binance candles: ", "error", err)
			mu.Lock()
			errs = append(errs, fmt.Errorf("Binance candles: %w", err))
			mu.Unlock()
			return
		}

		if len(binanceInstruments.Symbols) == 0 || len(ohlc) == 0 {
			mu.Lock()
			errs = append(errs, fmt.Errorf("Binance: no data available"))
			mu.Unlock()
			return
		}

		mu.Lock()
		result = append(result, dto.SearchInstrumentResponse{
			ExchangeName:   "Binance",
			Symbol:         binanceInstruments.Symbols[0].Symbol,
			LogoURL:        fmt.Sprintf("https://img.logo.dev/crypto/%s?token=%s", strings.ToLower(before), os.Getenv("DEV_LOGO_PK")),
			OHLCLastCandle: ohlc[len(ohlc)-1],
		})
		mu.Unlock()
	}()

	wg.Wait()

	if len(errs) > 0 {
		for _, err := range errs {
			s.Log.Error("error in search instruments: ", "error", err)
		}
	}

	if len(result) == 0 {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch data from all exchanges"})
		return
	}

	ctx.JSON(http.StatusOK, result)
}

func (s *ServiceExchange) GetCandles(ctx *gin.Context) {
	instrument := ctx.Query("instrument")
	if instrument == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "instrument is required"})
		return
	}

	interval := ctx.Query("interval")
	after := ctx.Query("after")

}
