package servicehttp

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Flikest/PingVi_backend/internal/delivery/dto"
	"github.com/Flikest/PingVi_backend/pkg/alchemy"
	"github.com/Flikest/PingVi_backend/pkg/gcm"
	"github.com/Flikest/PingVi_backend/pkg/mempool"

	"github.com/Flikest/PingVi_backend/pkg/tokens"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (s *ServiceWallet) sendOkxTransaction(
	apiKey string,
	apiSecret string,
	passphrase string,
	ccy string,
	chain string,
	toAddr string,
	amt string,
	dest string,
) (*dto.WithdrawOkxResponse, error) {
	const (
		baseURL     = "https://www.okx.com"
		endpoint    = "/api/v5/asset/withdrawal"
		httpTimeout = 30 * time.Second
	)

	reqBody := dto.WithdrawOkxRequest{
		Amt:    amt,
		Dest:   dest,
		Ccy:    ccy,
		Chain:  chain,
		ToAddr: toAddr,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request body: %w", err)
	}

	timestamp := time.Now().UTC().Format("2006-01-02T15:04:05.000Z")

	signPayload := timestamp + http.MethodPost + endpoint + string(bodyBytes)

	mac := hmac.New(sha256.New, []byte(apiSecret))
	mac.Write([]byte(signPayload))
	signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	req, err := http.NewRequest(http.MethodPost, baseURL+endpoint, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("OK-ACCESS-KEY", apiKey)
	req.Header.Set("OK-ACCESS-SIGN", signature)
	req.Header.Set("OK-ACCESS-TIMESTAMP", timestamp)
	req.Header.Set("OK-ACCESS-PASSPHRASE", passphrase)

	client := &http.Client{Timeout: httpTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(respBytes))
	}

	var withdrawResp dto.WithdrawOkxResponse
	if err := json.Unmarshal(respBytes, &withdrawResp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	if withdrawResp.Code != "0" {
		return &withdrawResp, fmt.Errorf("API error: code=%s, msg=%s", withdrawResp.Code, withdrawResp.Msg)
	}

	return &withdrawResp, nil
}

func (s *ServiceWallet) sendByBitTransaction(
	apiKey string,
	apiSecret string,
	coin string,
	chain string,
	address string,
	amount string,
	accountType string,
	forceChain int,
) (*dto.WithdrawByBitResponse, error) {
	const (
		baseURL     = "https://api-testnet.bybit.com"
		endpoint    = "/v5/asset/withdraw/create"
		recvWindow  = "5000"
		httpTimeout = 30 * time.Second
	)

	reqBody := dto.WithdrawByBitRequest{
		Coin:        coin,
		Chain:       chain,
		Address:     address,
		Amount:      amount,
		Timestamp:   time.Now().UnixMilli(),
		ForceChain:  forceChain,
		AccountType: accountType,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request body: %w", err)
	}

	timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)

	signPayload := timestamp + apiKey + recvWindow + string(bodyBytes)

	mac := hmac.New(sha256.New, []byte(apiSecret))
	mac.Write([]byte(signPayload))
	signature := hex.EncodeToString(mac.Sum(nil))

	req, err := http.NewRequest(http.MethodPost, baseURL+endpoint, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-BAPI-API-KEY", apiKey)
	req.Header.Set("X-BAPI-TIMESTAMP", timestamp)
	req.Header.Set("X-BAPI-RECV-WINDOW", recvWindow)
	req.Header.Set("X-BAPI-SIGN", signature)

	client := &http.Client{Timeout: httpTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(respBytes))
	}

	var withdrawResp dto.WithdrawByBitResponse
	if err := json.Unmarshal(respBytes, &withdrawResp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	if withdrawResp.RetCode != 0 {
		return &withdrawResp, fmt.Errorf("API error: code=%d, msg=%s", withdrawResp.RetCode, withdrawResp.RetMsg)
	}

	return &withdrawResp, nil
}

func (s *ServiceWallet) getCeFiBalance(ctx context.Context, userID uuid.UUID) (BalanceResponse, error) {
	bybitKeys, err := s.Repository.SelectExchangeApiKeys(ctx, userID, "bybit")
	if err != nil {
		s.Log.Error("error with select bybit api keys: ", "error", err)
	}
	okxKeys, err := s.Repository.SelectExchangeApiKeys(ctx, userID, "okx")
	if err != nil {
		s.Log.Error("error with select okx api keys: ", "error", err)
		return BalanceResponse{}, err
	}

	deks := struct{ byBitDek, okxDek string }{}

	if bybitKeys.EncryptedDek != "" {
		dek, err := s.KMS.Decrypt(bybitKeys.EncryptedDek, "dek_crypto")
		if err != nil {
			s.Log.Error("error with decrypt bybit dek: ", "error", err)
			return BalanceResponse{}, err
		} else {
			deks.byBitDek = dek
		}
	}
	if okxKeys.EncryptedDek != "" {
		dek, err := s.KMS.Decrypt(okxKeys.EncryptedDek, "dek_crypto")
		if err != nil {
			s.Log.Error("error with decrypt okx dek: ", "error", err)
			return BalanceResponse{}, err
		} else {
			deks.okxDek = dek
		}
	}

	result := BalanceResponse{Coins: []CoinQuote{}}
	coinsBySymbol := map[string]*CoinQuote{}
	client := &http.Client{Timeout: 10 * time.Second}

	if bybitKeys.EncryptedApiKey != "" {
		apikey, err := gcm.Decrypt([]byte(bybitKeys.EncryptedApiKey), []byte(deks.byBitDek))
		if err != nil {
			s.Log.Error("error decrypt bybit apikey: ", "error", err)
			return BalanceResponse{}, err
		} else {
			secretKey, err := gcm.Decrypt([]byte(bybitKeys.EncryptedSecretKey), []byte(deks.byBitDek))
			if err != nil {
				s.Log.Error("error decrypt bybit secret: ", "error", err)
				return BalanceResponse{}, err
			} else {
				timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)
				recvWindow := "5000"
				queryString := "accountType=UNIFIED"

				paramStr := timestamp + string(apikey) + recvWindow + queryString
				mac := hmac.New(sha256.New, []byte(secretKey))
				mac.Write([]byte(paramStr))
				signature := hex.EncodeToString(mac.Sum(nil))

				req, _ := http.NewRequest("GET", "https://api.bybit.com/v5/account/wallet-balance?"+queryString, nil)
				req.Header.Add("X-BAPI-API-KEY", string(apikey))
				req.Header.Add("X-BAPI-TIMESTAMP", timestamp)
				req.Header.Add("X-BAPI-RECV-WINDOW", recvWindow)
				req.Header.Add("X-BAPI-SIGN", signature)
				req.Header.Add("X-BAPI-SIGN-TYPE", "2")

				res, err := client.Do(req)
				if err != nil {
					s.Log.Error("bybit request failed: ", "error", err)
					return BalanceResponse{}, err
				} else {
					body, _ := io.ReadAll(res.Body)
					res.Body.Close()

					var resp ByBitWalletRestResponse
					if err := json.Unmarshal(body, &resp); err != nil {
						s.Log.Error("bybit unmarshal: ", "error", err)
						return BalanceResponse{}, err
					} else if resp.RetCode == 0 && len(resp.Result.List) > 0 {
						acc := resp.Result.List[0]
						result.TotalEquity = addDecimal(result.TotalEquity, acc.TotalEquity)
						result.TotalPerpUPL = addDecimal(result.TotalPerpUPL, acc.TotalPerpUPL)

						for _, c := range acc.Coin {
							mergeCoin(coinsBySymbol, CoinQuote{
								ExchangeName:  "ByBit",
								Coin:          fmt.Sprintf("%s/usdt", c.Coin),
								Icon:          fmt.Sprintf("https://img.logokit.com/crypto/%s?token=pk_fr477644f9651bfdf397b7", strings.ToUpper(c.Coin)),
								WalletBalance: c.WalletBalance,
								UsdValue:      c.UsdValue,
								UnrealisedPnl: c.UnrealisedPnl,
							})
						}
					} else {
						s.Log.Error("bybit api error: ", "retCode", resp.RetCode, "retMsg", resp.RetMsg)
						return BalanceResponse{}, errors.New("bybit api error")
					}
				}
			}
		}
	}

	if okxKeys.EncryptedApiKey != "" {
		apikey, err := gcm.Decrypt([]byte(okxKeys.EncryptedApiKey), []byte(deks.okxDek))
		if err != nil {
			s.Log.Error("error decrypt okx apikey: ", "error", err)
		} else {
			secretKey, err := gcm.Decrypt([]byte(okxKeys.EncryptedSecretKey), []byte(deks.okxDek))
			if err != nil {
				s.Log.Error("error decrypt okx secret: ", "error", err)
			} else {
				passphrase, err := gcm.Decrypt([]byte(okxKeys.EncryptedApiKey), []byte(deks.okxDek))
				if err != nil {
					s.Log.Error("error decrypt okx passphrase: ", "error", err)
				} else {
					timestamp := time.Now().UTC().Format("2006-01-02T15:04:05.000Z")
					requestPath := "/api/v5/account/balance"

					prehash := timestamp + "GET" + requestPath
					mac := hmac.New(sha256.New, []byte(secretKey))
					mac.Write([]byte(prehash))
					signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))

					req, _ := http.NewRequest("GET", "https://openapi.okx.com"+requestPath, nil)
					req.Header.Add("OK-ACCESS-KEY", string(apikey))
					req.Header.Add("OK-ACCESS-SIGN", signature)
					req.Header.Add("OK-ACCESS-TIMESTAMP", timestamp)
					req.Header.Add("OK-ACCESS-PASSPHRASE", string(passphrase))
					req.Header.Add("Content-Type", "application/json")

					res, err := client.Do(req)
					if err != nil {
						s.Log.Error("okx request failed: ", "error", err)
					} else {
						body, _ := io.ReadAll(res.Body)
						res.Body.Close()

						var resp OKXBalanceResponse
						if err := json.Unmarshal(body, &resp); err != nil {
							s.Log.Error("okx unmarshal: ", "error", err)
						} else if resp.Code == "0" && len(resp.Data) > 0 {
							d := resp.Data[0]
							result.TotalEquity = addDecimal(result.TotalEquity, d.TotalEq)
							result.TotalPerpUPL = addDecimal(result.TotalPerpUPL, d.Upl)

							for _, c := range d.Details {
								mergeCoin(coinsBySymbol, CoinQuote{
									ExchangeName:  "OKX",
									Coin:          fmt.Sprintf("%s/usdt", c.Ccy),
									WalletBalance: c.CashBal,
									UsdValue:      c.EqUsd,
									UnrealisedPnl: c.Upl,
									AvgEntryPrice: c.AvgPx,
								})
							}
						} else {
							s.Log.Error("okx api error: ", "code", resp.Code, "msg", resp.Msg)
						}
					}
				}
			}
		}
	}

	for _, c := range coinsBySymbol {
		result.Coins = append(result.Coins, *c)
	}

	return result, nil
}

func (s *ServiceWallet) getDeFiBalance(ctx context.Context, userID uuid.UUID) (BalanceResponse, error) {
	const maxRetries = 5
	const baseDelay = 1 * time.Second

	var balance BalanceResponse

	keys, err := s.Repository.SelectPublicKeys(ctx, userID)
	if err != nil {
		s.Log.Error("error with select public keys: ", "error", err)
		return BalanceResponse{}, err
	}

	if keys.Btc != "" {
		btcUsd, err := mempool.GetBtcPrices()
		if err != nil {
			s.Log.Error("error with getting btc price: ", "error", err)
			return BalanceResponse{}, err
		}

		btcBalance, err := mempool.GetBtcBalance(keys.Btc)
		if err != nil {
			s.Log.Error("error with getting btc balance: ", "error", err)
			return BalanceResponse{}, err
		}

		balance.Coins = append(balance.Coins, CoinQuote{
			ExchangeName:  "",
			Coin:          "btc/usd",
			UsdValue:      fmt.Sprintf("%f", btcBalance*float64(btcUsd)),
			WalletBalance: fmt.Sprintf("%f", btcBalance),
		})

		balance.TotalEquity += fmt.Sprintf("%f", btcBalance*float64(btcUsd))
	}

	if keys.Eth != "" && keys.Sol != "" {
		for attempt := 0; attempt <= maxRetries; attempt++ {
			ID, key, err := s.Repository.SelectActiveAlchemyKey(ctx)
			if err != nil {
				s.Log.Error("error with selecting alchemy keys: ", "error", err)
				return BalanceResponse{}, err
			}

			decryptApiKey, err := s.KMS.Decrypt(key, "alchemy_ctypto")
			if err != nil {
				s.Log.Error("error with decrypting alchemy api key: ", "error", err)
				return BalanceResponse{}, err
			}

			evmSolanaBalance, StatusCode, err := alchemy.GetTokenBalances(decryptApiKey, keys.Eth, keys.Sol)
			if err != nil {
				s.Log.Error("error with getting token balance: ", "error", err)
				return BalanceResponse{}, err
			}

			if StatusCode == http.StatusTooManyRequests {
				s.Log.Info("many requests to alchemy api on ID: ", "ID", ID)

				if err := s.Repository.UpdateTokenStatus(ctx, ID); err != nil {
					s.Log.Error("error with updating status token: ", "error", err)
					return BalanceResponse{}, err
				}

				if attempt >= maxRetries {
					s.Log.Error("max retries exceeded for alchemy API")
					return BalanceResponse{}, fmt.Errorf("max retries exceeded, status code: %d", StatusCode)
				}

				delay := baseDelay * time.Duration(1<<uint(attempt))
				s.Log.Info(fmt.Sprintf("retrying after %v, attempt: %d", delay, attempt+1))

				select {
				case <-ctx.Done():
					return BalanceResponse{}, ctx.Err()
				case <-time.After(delay):
				}

				continue
			}

			for _, token := range evmSolanaBalance {
				balance.TotalEquity = token.PriceUSD

				balance.Coins = append(balance.Coins, CoinQuote{
					ExchangeName:  "",
					Coin:          fmt.Sprintf("%s/usd", strings.ToUpper(token.Symbol)),
					Icon:          fmt.Sprintf("https://img.logokit.com/crypto/%s?token=pk_fr477644f9651bfdf397b7", strings.ToUpper(token.Symbol)),
					UsdValue:      token.PriceUSD,
					WalletBalance: token.Balance,
				})
			}
			break
		}
	}

	return balance, nil
}

func (s *ServiceWallet) GetBalance(ctx *gin.Context) {
	payload, err := tokens.Verify(ctx.Request.Header.Get("Authorization"), []byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		s.Log.Error("error with verify jwt user token: ", "error", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, err := uuid.Parse(payload.ID.String())
	if err != nil {
		s.Log.Error("error with parsing jwt token from header: ", "error", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	accountType := ctx.Param("account_type")

	balance := struct {
		data BalanceResponse
	}{}

	switch accountType {
	case "defi":
		result, err := s.getDeFiBalance(ctx.Request.Context(), userID)
		if err != nil {
			s.Log.Error("error with get defi user balance: ", "error", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch balance"})
			return
		}
		balance.data = result
	case "cefi":
		result, err := s.getCeFiBalance(ctx.Request.Context(), userID)
		if err != nil {
			s.Log.Error("error with get cefi user balance: ", "error", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch balance"})
			return
		}
		balance.data = result
	}

	totalEquity, err := strconv.ParseFloat(balance.data.TotalEquity, 64)
	if err != nil {
		s.Log.Error("error with parse float 64: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	if err := s.Repository.ActualizeSnapshot(ctx.Request.Context(), userID, "cefi", totalEquity, time.Now()); err != nil {
		s.Log.Error("error with actialize snapshot: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	ctx.JSON(http.StatusOK, balance.data)
}

func (s *ServiceWallet) GetTransactions(ctx *gin.Context) {
	payload, err := tokens.Verify(ctx.Request.Header.Get("Authorization"), []byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		s.Log.Error("error with verify jwt user token: ", "error", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, err := uuid.Parse(payload.ID.String())
	if err != nil {
		s.Log.Error("error with parsing jwt token from header: ", "error", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	bybitKeys, err := s.Repository.SelectExchangeApiKeys(ctx.Request.Context(), userID, "bybit")
	if err != nil {
		s.Log.Error("error with select bybit api keys: ", "error", err)
	}
	okxKeys, err := s.Repository.SelectExchangeApiKeys(ctx.Request.Context(), userID, "okx")
	if err != nil {
		s.Log.Error("error with select okx api keys: ", "error", err)
	}

	deks := struct{ byBitDek, okxDek string }{}

	if bybitKeys.EncryptedDek != "" {
		dek, err := s.KMS.Decrypt(bybitKeys.EncryptedDek, "dek_crypto")
		if err != nil {
			s.Log.Error("error with decrypt bybit dek: ", "error", err)
		} else {
			deks.byBitDek = dek
		}
	}
	if okxKeys.EncryptedDek != "" {
		dek, err := s.KMS.Decrypt(okxKeys.EncryptedDek, "dek_crypto")
		if err != nil {
			s.Log.Error("error with decrypt okx dek: ", "error", err)
		} else {
			deks.okxDek = dek
		}
	}

	var transactions []Transactions
	client := &http.Client{Timeout: 10 * time.Second}

	if bybitKeys.EncryptedApiKey != "" {
		apikey, err := gcm.Decrypt([]byte(bybitKeys.EncryptedApiKey), []byte(deks.byBitDek))
		if err != nil {
			s.Log.Error("error decrypt bybit apikey: ", "error", err)
		} else {
			secretKey, err := gcm.Decrypt([]byte(bybitKeys.EncryptedSecretKey), []byte(deks.byBitDek))
			if err != nil {
				s.Log.Error("error decrypt bybit secret: ", "error", err)
			} else {
				timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)
				recvWindow := "5000"
				queryString := "accountType=UNIFIED&limit=50"

				paramStr := timestamp + string(apikey) + recvWindow + queryString
				mac := hmac.New(sha256.New, []byte(secretKey))
				mac.Write([]byte(paramStr))
				signature := hex.EncodeToString(mac.Sum(nil))

				req, _ := http.NewRequest("GET", "https://api.bybit.com/v5/account/transaction-log?"+queryString, nil)
				req.Header.Add("X-BAPI-API-KEY", string(apikey))
				req.Header.Add("X-BAPI-TIMESTAMP", timestamp)
				req.Header.Add("X-BAPI-RECV-WINDOW", recvWindow)
				req.Header.Add("X-BAPI-SIGN", signature)
				req.Header.Add("X-BAPI-SIGN-TYPE", "2")

				res, err := client.Do(req)
				if err != nil {
					s.Log.Error("bybit transaction-log request failed: ", "error", err)
				} else {
					body, _ := io.ReadAll(res.Body)
					res.Body.Close()

					var resp ByBitTransactionLogResponse
					if err := json.Unmarshal(body, &resp); err != nil {
						s.Log.Error("bybit transaction-log unmarshal: ", "error", err)
					} else if resp.RetCode == 0 {
						for _, t := range resp.Result.List {

							timestamp, err := strconv.ParseInt(t.TransactionTime, 10, 64)
							if err != nil {
								s.Log.Error("error with perse int: ", "error", err)
								ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error with perse transaction time"})
							}

							transactions = append(transactions, Transactions{
								ID:        t.ID,
								Exchange:  "bybit",
								Symbol:    t.Symbol,
								Side:      strings.ToLower(t.Side),
								Type:      strings.ToLower(t.Type),
								Price:     t.TradePrice,
								Quantity:  t.Qty,
								Fee:       t.Fee,
								Funding:   t.Funding,
								PnL:       t.CashFlow,
								Balance:   t.CashBalance,
								Timestamp: timestamp,
							})
						}
					} else {
						s.Log.Error("bybit transaction-log api error: ", "retCode", resp.RetCode, "retMsg", resp.RetMsg)
					}
				}
			}
		}
	}

	// ============ OKX ============
	if okxKeys.EncryptedApiKey != "" {
		apikey, err := gcm.Decrypt([]byte(okxKeys.EncryptedApiKey), []byte(deks.okxDek))
		if err != nil {
			s.Log.Error("error decrypt okx apikey: ", "error", err)
		} else {
			secretKey, err := gcm.Decrypt([]byte(okxKeys.EncryptedSecretKey), []byte(deks.okxDek))
			if err != nil {
				s.Log.Error("error decrypt okx secret: ", "error", err)
			} else {
				passphrase, err := gcm.Decrypt([]byte(okxKeys.EncryptedApiKey), []byte(deks.okxDek))
				if err != nil {
					s.Log.Error("error decrypt okx passphrase: ", "error", err)
				} else {
					timestamp := time.Now().UTC().Format("2006-01-02T15:04:05.000Z")
					requestPath := "/api/v5/account/bills"
					queryString := "?limit=100"
					body := ""

					prehash := timestamp + "GET" + requestPath + queryString + body
					mac := hmac.New(sha256.New, []byte(secretKey))
					mac.Write([]byte(prehash))
					signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))

					req, _ := http.NewRequest("GET", "https://openapi.okx.com"+requestPath+queryString, nil)
					req.Header.Add("OK-ACCESS-KEY", string(apikey))
					req.Header.Add("OK-ACCESS-SIGN", signature)
					req.Header.Add("OK-ACCESS-TIMESTAMP", timestamp)
					req.Header.Add("OK-ACCESS-PASSPHRASE", string(passphrase))
					req.Header.Add("Content-Type", "application/json")

					res, err := client.Do(req)
					if err != nil {
						s.Log.Error("okx bills request failed: ", "error", err)
					} else {
						body, _ := io.ReadAll(res.Body)
						res.Body.Close()

						var resp OKXBillsResponse
						if err := json.Unmarshal(body, &resp); err != nil {
							s.Log.Error("okx bills unmarshal: ", "error", err)
						} else if resp.Code == "0" {
							for _, t := range resp.Data {

								timestamp, err := strconv.ParseInt(t.Ts, 10, 64)
								if err != nil {
									s.Log.Error("error with perse int: ", "error", err)
									ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error with perse transaction time"})
								}

								transactions = append(transactions, Transactions{
									ID:        t.BillId,
									Exchange:  "okx",
									Symbol:    t.InstId,
									Side:      mapOKXSide(t.Type),
									Type:      mapOKXType(t.Type),
									Price:     t.Px,
									Quantity:  t.Sz,
									PnL:       t.BalChg,
									Balance:   t.Bal,
									Timestamp: timestamp,
								})
							}
						} else {
							s.Log.Error("okx bills api error: ", "code", resp.Code, "msg", resp.Msg)
						}
					}
				}
			}
		}
	}

	ctx.JSON(http.StatusOK, transactions)
}

func (s *ServiceWallet) CreateWallet(ctx *gin.Context) {
	payload, err := tokens.Verify(ctx.Request.Header.Get("Authorization"), []byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		s.Log.Error("error with verify jwt user token: ", "error", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, err := uuid.Parse(payload.ID.String())
	if err != nil {
		s.Log.Error("error with parsing jwt token from header: ", "error", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var body dto.CreateWallet
	if err := ctx.BindJSON(&body); err != nil {
		s.Log.Error("invalid body: ", "error", err)
		ctx.JSON(http.StatusBadRequest, err)
		return
	}

	createdAt := time.Now()

	if err := s.Repository.CreateWallet(ctx.Request.Context(), userID, body, createdAt); err != nil {
		s.Log.Error("error with creating wallet: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	ctx.JSON(http.StatusCreated, body)
}

func (s *ServiceWallet) GetSnapshot(ctx *gin.Context) {
	payload, err := tokens.Verify(ctx.Request.Header.Get("Authorization"), []byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		s.Log.Error("error with verify jwt user token: ", "error", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, err := uuid.Parse(payload.ID.String())
	if err != nil {
		s.Log.Error("error with parsing jwt token from header: ", "error", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	data, err := s.Repository.SelectSnapshotsByUserID(ctx.Request.Context(), userID)
	if err != nil {
		s.Log.Error("error with getting snapshot: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	ctx.JSON(http.StatusOK, data)
}

func (s *ServiceWallet) SendTransaction(ctx *gin.Context) {
	payload, err := tokens.Verify(ctx.Request.Header.Get("Authorization"), []byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		s.Log.Error("error with verify jwt user token: ", "error", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, err := uuid.Parse(payload.ID.String())
	if err != nil {
		s.Log.Error("error with parsing jwt token from header: ", "error", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var body dto.SendTransaction
	if err := ctx.BindJSON(&body); err != nil {
		s.Log.Error("invalid body in send transaction")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	switch body.Type {
	case "defi":
		switch body.Net {
		case "bitcoin":
			status, err := mempool.SendTransaction(body.TXhex)
			if err != nil {
				ctx.JSON(status, gin.H{"error": err})
				return
			}

			ctx.JSON(status, gin.H{"status": "success"})
		default:
			const maxRetries = 5
			const baseDelay = 1 * time.Second

			for attempt := 0; attempt <= maxRetries; attempt++ {
				ID, apiKey, err := s.Repository.SelectActiveAlchemyKey(ctx.Request.Context())
				if err != nil {
					s.Log.Error("error with selecting alchemy key: ", "error", err)
					ctx.JSON(http.StatusInternalServerError, gin.H{"error": "no active alchemy key"})
					return
				}

				decryptApiKey, err := s.KMS.Decrypt(apiKey, "alchemy_ctypto")
				if err != nil {
					s.Log.Error("error with decrypting alchemy api key: ", "error", err)
					ctx.JSON(http.StatusInternalServerError, gin.H{"error": err})
					return
				}

				status, err := alchemy.SendTransaction(body.Net, decryptApiKey, body.TXhex)
				if err != nil {
					s.Log.Error("error with sending transaction: ", "error", err)
					ctx.JSON(status, gin.H{"error": err})
					return
				}

				if status == http.StatusTooManyRequests {
					s.Log.Info("many requests to alchemy api on ID: ", "ID", ID)

					if err := s.Repository.UpdateTokenStatus(ctx.Request.Context(), ID); err != nil {
						s.Log.Error("error with updating status token: ", "error", err)
						ctx.JSON(http.StatusInternalServerError, gin.H{"error": err})
						return
					}

					if attempt >= maxRetries {
						s.Log.Error("max retries exceeded for alchemy API")
						ctx.JSON(http.StatusTooManyRequests, gin.H{"error": "max retries exceeded, too many requests"})
						return
					}

					delay := baseDelay * time.Duration(1<<uint(attempt))
					s.Log.Info(fmt.Sprintf("retrying SendTransaction after %v, attempt: %d", delay, attempt+1))

					select {
					case <-ctx.Request.Context().Done():
						s.Log.Error("context cancelled during retry: ", "error", ctx.Request.Context().Err())
						ctx.JSON(http.StatusRequestTimeout, gin.H{"error": ctx.Request.Context().Err()})
						return
					case <-time.After(delay):
					}

					continue
				}

				ctx.JSON(status, gin.H{"status": "success"})
				return
			}
		}
	case "cefi":
		switch body.ExchangeName {
		case "okx":
			okxKeys, err := s.Repository.SelectExchangeApiKeys(ctx.Request.Context(), userID, body.ExchangeName)
			if err != nil {
				s.Log.Error("error with selecting api keys: ", "error", err)
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": err})
				return
			}

			var dek struct {
				dek string
			}

			if okxKeys.EncryptedDek != "" {
				okxDek, err := s.KMS.Decrypt(okxKeys.EncryptedDek, "dek_crypto")
				if err != nil {
					s.Log.Error("error with decrypt bybit dek: ", "error", err)
					ctx.JSON(http.StatusInternalServerError, gin.H{"error": err})
					return
				}
				dek.dek = okxDek
			}

			apiKey, err := gcm.Decrypt([]byte(okxKeys.EncryptedApiKey), []byte(dek.dek))
			if err != nil {
				s.Log.Error("error with decrypting api key: ", "error", err)
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": err})
				return
			}
			secretKey, err := gcm.Decrypt([]byte(okxKeys.EncryptedSecretKey), []byte(dek.dek))
			if err != nil {
				s.Log.Error("error with dectypting secret key: ", "error", err)
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": err})
				return
			}
			passphrase, err := gcm.Decrypt([]byte(okxKeys.EnrcyptedPassphrase), []byte(dek.dek))
			if err != nil {
				s.Log.Error("error with decrypting dek: ", "error", err)
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": err})
				return
			}

			withdrawResp, err := s.sendOkxTransaction(apiKey, secretKey, passphrase, body.CCY, body.Chain, body.ToAddress, body.Amount, body.Dest)
			if err != nil {
				s.Log.Error("error with send okx transaction: ", "error", err)
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": err})
				return
			}

			ctx.JSON(http.StatusOK, withdrawResp)
			return
		case "bybit":
			bybitKeys, err := s.Repository.SelectExchangeApiKeys(ctx.Request.Context(), userID, body.ExchangeName)
			if err != nil {
				s.Log.Error("error with selecting api keys: ", "error", err)
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": err})
				return
			}

			var dek struct {
				dek string
			}

			if bybitKeys.EncryptedDek != "" {
				bybitDek, err := s.KMS.Decrypt(bybitKeys.EncryptedDek, "dek_crypto")
				if err != nil {
					s.Log.Error("error with decryptring bybit dek: ", "error", err)
					ctx.JSON(http.StatusInternalServerError, gin.H{"error": err})
					return
				}
				dek.dek = bybitDek
			}

			apiKey, err := gcm.Decrypt([]byte(bybitKeys.EncryptedApiKey), []byte(dek.dek))
			if err != nil {
				s.Log.Error("error with decrypting api key: ", "error", err)
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": err})
				return
			}
			secretKey, err := gcm.Decrypt([]byte(bybitKeys.EncryptedSecretKey), []byte(dek.dek))
			if err != nil {
				s.Log.Error("error with dectypting secret key: ", "error", err)
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": err})
				return
			}

			withdrawResp, err := s.sendByBitTransaction(apiKey, secretKey, body.CCY, body.Chain, body.ToAddress, body.Amount, "FUND", 0)
			if err != nil {
				s.Log.Error("error with send bybit transaction: ", "error", err)
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": err})
				return
			}
			ctx.JSON(http.StatusOK, withdrawResp)
		default:
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "unsupported exchange"})
		}
	default:
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "unsupported account type"})
	}
}

func (s *ServiceWallet) GetListCurrenciesConversion(ctx *gin.Context) {
	payload, err := tokens.Verify(ctx.Request.Header.Get("Authorization"), []byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		s.Log.Error("error with verify jwt user token: ", "error", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, err := uuid.Parse(payload.ID.String())
	if err != nil {
		s.Log.Error("error with parsing jwt token from header: ", "error", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	exchangeName := ctx.Param("exchange_name")

	var result []dto.ConvertCurrency

	switch exchangeName {
	case "okx":
		url := "https://openapi.okx.com/api/v5/asset/convert/currencies"

		resp, err := http.Get(url)
		if err != nil {
			s.Log.Error("error with getting okx convert currencies: ", "error", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}
		defer resp.Body.Close()

		var okxconvertCurrencies dto.OKXConvertCurrencies

		if err := json.NewDecoder(resp.Body).Decode(&okxconvertCurrencies); err != nil {
			s.Log.Error("error with decoding okx convert currencies: ", "error", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}

		for _, data := range okxconvertCurrencies.Data {
			result = append(result, dto.ConvertCurrency{
				CCY:  data.Ccy,
				Min:  data.Min,
				Max:  data.Max,
				Icon: fmt.Sprintf("https://img.logokit.com/crypto/%s?token=pk_fr477644f9651bfdf397b7", strings.ToUpper(data.Ccy)),
			})
		}

	case "bybit":
		bybitKeys, err := s.Repository.SelectExchangeApiKeys(ctx.Request.Context(), userID, exchangeName)
		if err != nil {
			s.Log.Error("error with selecting api keys: ", "error", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}

		var dek struct {
			dek string
		}

		if bybitKeys.EncryptedDek != "" {
			bybitDek, err := s.KMS.Decrypt(bybitKeys.EncryptedDek, "dek_crypto")
			if err != nil {
				s.Log.Error("error with decrypting bybit dek: ", "error", err)
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": err})
				return
			}
			dek.dek = bybitDek
		}

		apiKey, err := gcm.Decrypt([]byte(bybitKeys.EncryptedApiKey), []byte(dek.dek))
		if err != nil {
			s.Log.Error("error with decrypting api key: ", "error", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}
		secretKey, err := gcm.Decrypt([]byte(bybitKeys.EncryptedSecretKey), []byte(dek.dek))
		if err != nil {
			s.Log.Error("error with decrypting secret key: ", "error", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}

		const (
			baseURL    = "https://api.bybit.com"
			recvWindow = "5000"
		)

		q := url.Values{}
		q.Set("accountType", "eb_convert_funding")
		q.Set("side", "0")
		queryString := q.Encode()

		ts := strconv.FormatInt(time.Now().UnixMilli(), 10)
		signPayload := ts + string(apiKey) + recvWindow + queryString

		mac := hmac.New(sha256.New, []byte(secretKey))
		mac.Write([]byte(signPayload))
		sign := hex.EncodeToString(mac.Sum(nil))

		req, err := http.NewRequest(http.MethodGet, baseURL+"/v5/asset/exchange/query-coin-list?"+queryString, nil)
		if err != nil {
			s.Log.Error("error with building bybit request: ", "error", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}
		req.Header.Set("X-BAPI-API-KEY", string(apiKey))
		req.Header.Set("X-BAPI-TIMESTAMP", ts)
		req.Header.Set("X-BAPI-RECV-WINDOW", recvWindow)
		req.Header.Set("X-BAPI-SIGN", sign)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			s.Log.Error("error with getting bybit convert currencies: ", "error", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}
		defer resp.Body.Close()

		var bybitResp dto.BybitConvertCoinListResponse
		if err := json.NewDecoder(resp.Body).Decode(&bybitResp); err != nil {
			s.Log.Error("error with decoding bybit convert currencies: ", "error", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}

		if bybitResp.RetCode != 0 {
			s.Log.Error("bybit retCode: ", "retCode", bybitResp.RetCode, "retMsg", bybitResp.RetMsg)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": bybitResp.RetMsg})
			return
		}

		for _, data := range bybitResp.Result.Coins {
			if data.DisableFrom {
				continue
			}
			result = append(result, dto.ConvertCurrency{
				CCY:  data.Coin,
				Min:  data.SingleFromMinLimit,
				Max:  data.SingleFromMaxLimit,
				Icon: fmt.Sprintf("https://img.logokit.com/crypto/%s?token=pk_fr477644f9651bfdf397b7", strings.ToUpper(data.Coin)),
			})
		}

	default:
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "unsupported exchange"})
		return
	}

	ctx.JSON(http.StatusOK, result)
}

func (s *ServiceWallet) GetConvertQuote(ctx *gin.Context) {
	payload, err := tokens.Verify(ctx.Request.Header.Get("Authorization"), []byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		s.Log.Error("error with verify jwt user token: ", "error", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, err := uuid.Parse(payload.ID.String())
	if err != nil {
		s.Log.Error("error with parsing jwt token from header: ", "error", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var body dto.GetConvertQuote
	if err := ctx.BindJSON(&body); err != nil {
		s.Log.Error("invalid body in get convert quote: ", "error", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	switch body.ExchangeName {
	case "okx":
		okxKeys, err := s.Repository.SelectExchangeApiKeys(ctx.Request.Context(), userID, body.ExchangeName)
		if err != nil {
			s.Log.Error("error with selecting api keys: ", "error", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}

		var dek struct {
			dek string
		}

		if okxKeys.EncryptedDek != "" {
			okxDek, err := s.KMS.Decrypt(okxKeys.EncryptedDek, "dek_crypto")
			if err != nil {
				s.Log.Error("error with decrypting okx dek: ", "error", err)
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": err})
				return
			}
			dek.dek = okxDek
		}

		apiKey, err := gcm.Decrypt([]byte(okxKeys.EncryptedApiKey), []byte(dek.dek))
		if err != nil {
			s.Log.Error("error with decrypting api key: ", "error", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}
		secretKey, err := gcm.Decrypt([]byte(okxKeys.EncryptedSecretKey), []byte(dek.dek))
		if err != nil {
			s.Log.Error("error with decrypting secret key: ", "error", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}
		passphrase, err := gcm.Decrypt([]byte(okxKeys.EnrcyptedPassphrase), []byte(dek.dek))
		if err != nil {
			s.Log.Error("error with decrypting passphrase: ", "error", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}

		bodyBytes, err := json.Marshal(map[string]string{
			"baseCcy":  body.FromCoin,
			"quoteCcy": body.ToCoin,
			"side":     "sell",
			"rfqSz":    body.Amount,
			"rfqSzCcy": body.FromCoin,
		})
		if err != nil {
			s.Log.Error("error with marshal okx body: ", "error", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}
		bodyStr := string(bodyBytes)

		const (
			baseURL = "https://openapi.okx.com"
			path    = "/api/v5/asset/convert/estimate-quote"
		)

		ts := time.Now().UTC().Format("2006-01-02T15:04:05.000Z")
		signPayload := ts + "POST" + path + bodyStr

		mac := hmac.New(sha256.New, []byte(secretKey))
		mac.Write([]byte(signPayload))
		sign := base64.StdEncoding.EncodeToString(mac.Sum(nil))

		req, err := http.NewRequest(http.MethodPost, baseURL+path, strings.NewReader(bodyStr))
		if err != nil {
			s.Log.Error("error with building okx request: ", "error", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("OK-ACCESS-KEY", string(apiKey))
		req.Header.Set("OK-ACCESS-SIGN", sign)
		req.Header.Set("OK-ACCESS-TIMESTAMP", ts)
		req.Header.Set("OK-ACCESS-PASSPHRASE", string(passphrase))

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			s.Log.Error("error with getting okx quote: ", "error", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}
		defer resp.Body.Close()

		var okxQuote dto.OKXEstimateQuoteResponse
		if err := json.NewDecoder(resp.Body).Decode(&okxQuote); err != nil {
			s.Log.Error("error with decoding okx quote: ", "error", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}

		if okxQuote.Code != "0" {
			s.Log.Error("okx code: ", "code", okxQuote.Code, "msg", okxQuote.Msg)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": okxQuote.Msg})
			return
		}

		ctx.JSON(http.StatusOK, okxQuote)
		return

	case "bybit":
		bybitKeys, err := s.Repository.SelectExchangeApiKeys(ctx.Request.Context(), userID, body.ExchangeName)
		if err != nil {
			s.Log.Error("error with selecting api keys: ", "error", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}

		var dek struct {
			dek string
		}

		if bybitKeys.EncryptedDek != "" {
			bybitDek, err := s.KMS.Decrypt(bybitKeys.EncryptedDek, "dek_crypto")
			if err != nil {
				s.Log.Error("error with decrypting bybit dek: ", "error", err)
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": err})
				return
			}
			dek.dek = bybitDek
		}

		apiKey, err := gcm.Decrypt([]byte(bybitKeys.EncryptedApiKey), []byte(dek.dek))
		if err != nil {
			s.Log.Error("error with decrypting api key: ", "error", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}
		secretKey, err := gcm.Decrypt([]byte(bybitKeys.EncryptedSecretKey), []byte(dek.dek))
		if err != nil {
			s.Log.Error("error with decrypting secret key: ", "error", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}

		bodyBytes, err := json.Marshal(map[string]string{
			"accountType":   "eb_convert_funding",
			"fromCoin":      body.FromCoin,
			"toCoin":        body.ToCoin,
			"requestCoin":   body.FromCoin,
			"requestAmount": body.Amount,
		})
		if err != nil {
			s.Log.Error("error with marshal bybit body: ", "error", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}
		bodyStr := string(bodyBytes)

		const (
			baseURL    = "https://api.bybit.com"
			recvWindow = "5000"
		)

		ts := strconv.FormatInt(time.Now().UnixMilli(), 10)
		signPayload := ts + string(apiKey) + recvWindow + bodyStr

		mac := hmac.New(sha256.New, []byte(secretKey))
		mac.Write([]byte(signPayload))
		sign := hex.EncodeToString(mac.Sum(nil))

		req, err := http.NewRequest(http.MethodPost, baseURL+"/v5/asset/exchange/quote-apply", strings.NewReader(bodyStr))
		if err != nil {
			s.Log.Error("error with building bybit request: ", "error", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-BAPI-API-KEY", string(apiKey))
		req.Header.Set("X-BAPI-TIMESTAMP", ts)
		req.Header.Set("X-BAPI-RECV-WINDOW", recvWindow)
		req.Header.Set("X-BAPI-SIGN", sign)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			s.Log.Error("error with getting bybit quote: ", "error", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}
		defer resp.Body.Close()

		var bybitQuote dto.BybitQuoteResponse
		if err := json.NewDecoder(resp.Body).Decode(&bybitQuote); err != nil {
			s.Log.Error("error with decoding bybit quote: ", "error", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}

		if bybitQuote.RetCode != 0 {
			s.Log.Error("bybit retCode: ", "retCode", bybitQuote.RetCode, "retMsg", bybitQuote.RetMsg)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": bybitQuote.RetMsg})
			return
		}

		ctx.JSON(http.StatusOK, bybitQuote)
		return

	default:
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "unsupported exchange"})
	}
}

func (s *ServiceWallet) ConfirmConvertQuote(ctx *gin.Context) {
	payload, err := tokens.Verify(ctx.Request.Header.Get("Authorization"), []byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		s.Log.Error("error with verify jwt user token: ", "error", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, err := uuid.Parse(payload.ID.String())
	if err != nil {
		s.Log.Error("error with parsing jwt token from header: ", "error", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var body dto.ConfirmConvertQuote
	if err := ctx.BindJSON(&body); err != nil {
		s.Log.Error("invalid body in confirm convert quote: ", "error", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	switch body.ExchangeName {
	case "bybit":
		bybitKeys, err := s.Repository.SelectExchangeApiKeys(ctx.Request.Context(), userID, body.ExchangeName)
		if err != nil {
			s.Log.Error("error with selecting api keys: ", "error", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}

		var dek struct {
			dek string
		}

		if bybitKeys.EncryptedDek != "" {
			bybitDek, err := s.KMS.Decrypt(bybitKeys.EncryptedDek, "dek_crypto")
			if err != nil {
				s.Log.Error("error with decrypting bybit dek: ", "error", err)
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": err})
				return
			}
			dek.dek = bybitDek
		}

		apiKey, err := gcm.Decrypt([]byte(bybitKeys.EncryptedApiKey), []byte(dek.dek))
		if err != nil {
			s.Log.Error("error with decrypting api key: ", "error", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}
		secretKey, err := gcm.Decrypt([]byte(bybitKeys.EncryptedSecretKey), []byte(dek.dek))
		if err != nil {
			s.Log.Error("error with decrypting secret key: ", "error", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}

		// Тело запроса: только quoteTxId
		bodyBytes, err := json.Marshal(map[string]string{
			"quoteTxId": body.QuoteTxId,
		})
		if err != nil {
			s.Log.Error("error with marshal bybit confirm body: ", "error", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}
		bodyStr := string(bodyBytes)

		const (
			baseURL    = "https://api.bybit.com"
			recvWindow = "5000"
		)

		ts := strconv.FormatInt(time.Now().UnixMilli(), 10)
		signPayload := ts + string(apiKey) + recvWindow + bodyStr

		mac := hmac.New(sha256.New, []byte(secretKey))
		mac.Write([]byte(signPayload))
		sign := hex.EncodeToString(mac.Sum(nil))

		req, err := http.NewRequest(http.MethodPost, baseURL+"/v5/asset/exchange/convert-execute", strings.NewReader(bodyStr))
		if err != nil {
			s.Log.Error("error with building bybit confirm request: ", "error", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-BAPI-API-KEY", string(apiKey))
		req.Header.Set("X-BAPI-TIMESTAMP", ts)
		req.Header.Set("X-BAPI-RECV-WINDOW", recvWindow)
		req.Header.Set("X-BAPI-SIGN", sign)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			s.Log.Error("error with confirming bybit quote: ", "error", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}
		defer resp.Body.Close()

		var bybitConfirm dto.BybitConfirmQuoteResponse
		if err := json.NewDecoder(resp.Body).Decode(&bybitConfirm); err != nil {
			s.Log.Error("error with decoding bybit confirm: ", "error", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}

		if bybitConfirm.RetCode != 0 {
			s.Log.Error("bybit confirm retCode: ", "retCode", bybitConfirm.RetCode, "retMsg", bybitConfirm.RetMsg)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": bybitConfirm.RetMsg})
			return
		}

		ctx.JSON(http.StatusOK, bybitConfirm)
		return

	default:
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "unsupported exchange"})
	}
}

func (s *ServiceWallet) AddExchangeApiKeys(ctx *gin.Context) {
	payload, err := tokens.Verify(ctx.Request.Header.Get("Authorization"), []byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		s.Log.Error("error with verify jwt user token: ", "error", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, err := uuid.Parse(payload.ID.String())
	if err != nil {
		s.Log.Error("error with parsing jwt token from header: ", "error", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var body dto.UserCexExchangeKeys
	if err := ctx.BindJSON(&body); err != nil {
		s.Log.Error("invalid body in add exchange api keys: ", "error", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	// Проверка, что все поля присутствуют
	if body.CexName == "" || body.EncryptedApiKey == "" || body.EncryptedSecretKey == "" || body.EnrcyptedPassphrase == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "exchange name, api key, secret and passphrase key are required"})
		return
	}

	// Шифруем api key и secret key
	dek := make([]byte, 32)
	if _, err := rand.Read(dek); err != nil {
		s.Log.Error("error with generating dek: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	encryptedDek, err := s.KMS.Encrypt(string(dek), "dek_crypto")
	if err != nil {
		s.Log.Error("error with encrypting dek: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	encryptedApiKey, err := gcm.Encrypt(dek, []byte(body.EncryptedApiKey))
	if err != nil {
		s.Log.Error("error with encrypting api key: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	encryptedSecretKey, err := gcm.Encrypt(dek, []byte(body.EncryptedSecretKey))
	if err != nil {
		s.Log.Error("error with encrypting secret key: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	encrypted_passphrase, err := gcm.Encrypt(dek, []byte(body.EnrcyptedPassphrase))
	if err != nil {
		s.Log.Error("error with encrypting passphrase key: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	body.UserID = userID.String()
	body.EnrcyptedPassphrase = encrypted_passphrase
	body.EncryptedDek = encryptedDek
	body.EncryptedApiKey = encryptedApiKey
	body.EncryptedSecretKey = encryptedSecretKey
	body.At = time.Now()

	if err := s.Repository.InsertExchangeApiKey(ctx.Request.Context(), body); err != nil {
		s.Log.Error("error with inserting api key: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"status": "success"})
}

func (s *ServiceWallet) UpdateExchangeApiKeys(ctx *gin.Context) {
	payload, err := tokens.Verify(ctx.Request.Header.Get("Authorization"), []byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		s.Log.Error("error with verify jwt user token: ", "error", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, err := uuid.Parse(payload.ID.String())
	if err != nil {
		s.Log.Error("error with parsing jwt token from header: ", "error", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var body dto.UserCexExchangeKeys
	if err := ctx.BindJSON(&body); err != nil {
		s.Log.Error("invalid body in add exchange api keys: ", "error", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	// Проверка, что все поля присутствуют
	if body.CexName == "" || body.EncryptedApiKey == "" || body.EncryptedSecretKey == "" || body.EnrcyptedPassphrase == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "exchange name, api key, secret and passphrase key are required"})
		return
	}

	// Шифруем api key и secret key
	dek := make([]byte, 32)
	if _, err := rand.Read(dek); err != nil {
		s.Log.Error("error with generating dek: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	encryptedDek, err := s.KMS.Encrypt(string(dek), "dek_crypto")
	if err != nil {
		s.Log.Error("error with encrypting dek: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	encryptedApiKey, err := gcm.Encrypt(dek, []byte(body.EncryptedApiKey))
	if err != nil {
		s.Log.Error("error with encrypting api key: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	encryptedSecretKey, err := gcm.Encrypt(dek, []byte(body.EncryptedSecretKey))
	if err != nil {
		s.Log.Error("error with encrypting secret key: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	encrypted_passphrase, err := gcm.Encrypt(dek, []byte(body.EnrcyptedPassphrase))
	if err != nil {
		s.Log.Error("error with encrypting passphrase key: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	body.UserID = userID.String()
	body.EnrcyptedPassphrase = encrypted_passphrase
	body.EncryptedDek = encryptedDek
	body.EncryptedApiKey = encryptedApiKey
	body.EncryptedSecretKey = encryptedSecretKey
	body.At = time.Now()

	if err := s.Repository.UpdateExchangeApiKeys(ctx.Request.Context(), body); err != nil {
		s.Log.Error("error with inserting api key: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "success"})
}

func (s *ServiceWallet) DeleteExchangeApiKeys(ctx *gin.Context) {
	payload, err := tokens.Verify(ctx.Request.Header.Get("Authorization"), []byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		s.Log.Error("error with verify jwt user token: ", "error", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, err := uuid.Parse(payload.ID.String())
	if err != nil {
		s.Log.Error("error with parsing jwt token from header: ", "error", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var body dto.DeleteApiKeys
	if err := ctx.BindJSON(&body); err != nil {
		s.Log.Error("invalid body in add exchange api keys: ", "error", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	if err := s.Repository.DeleteExchangeApiKeys(ctx.Request.Context(), userID, body.ExchangeName); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "success"})
}
