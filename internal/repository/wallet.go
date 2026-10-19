package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/Flikest/PingVi_backend/internal/delivery/dto"
	"github.com/google/uuid"
)

func (r *RepositoryWallet) CreateWallet(ctx context.Context, userID uuid.UUID, cw dto.CreateWallet, createdAt time.Time) error {
	query := `INSERT INTO wallets (user_id, eht_bnb_address, sol_address, btc_address, created_at) VALUES ($1, $2, $3, $4, $5)`

	_, err := r.DB.Exec(ctx, query,
		userID,
		cw.Eth.Address,
		cw.Sol.Address,
		cw.Btc.Address,
		createdAt,
	)
	if err != nil {
		r.Log.Error("error with creating wallet: ", "error", err)
		return err
	}

	return nil
}

func (r *RepositoryWallet) SelectPublicKeys(ctx context.Context, userID uuid.UUID) (dto.SelectKeys, error) {
	query := `SELECT eht_bnb_address, sol_address, btc_address, FROM wallets WHERE user_id = $1;`

	keys := dto.SelectKeys{}

	if err := r.DB.QueryRow(ctx, query, userID).Scan(&keys); err != nil {
		r.Log.Error("error with selecting walet public keys: ", "error", err)
		return dto.SelectKeys{}, err
	}

	return keys, nil
}

func (r *RepositoryWallet) SelectUserAddress(ctx context.Context, userID uuid.UUID, net string) (string, error) {
	query := `SELECT %s FROM wallets WHERE user_id = $1;`

	switch net {
	case "ethereum":
		query = fmt.Sprintf(query, "evm_address")
	case "solana":
		query = fmt.Sprintf(query, "sol_address")
	case "bitcoin":
		query = fmt.Sprintf(query, "btc_address")
	default:
		return "", fmt.Errorf("net is required field")
	}

	var address string

	if err := r.DB.QueryRow(ctx, query, userID).Scan(&address); err != nil {
		r.Log.Error("error with selecting user address by id and net: ", "error", err)
		return "", err
	}

	return address, nil
}

func (r *RepositoryWallet) InsertExchangeApiKey(ctx context.Context, userCexEchangeKeys dto.UserCexExchangeKeys) error {
	query := "INSERT INTO user_cex_exchange_keys(user_id, cex_name, encrypted_api_key, encrypted_secret_key, encrypted_passphrase, encrypted_dek) VALUES ($1, $2, $3, $4, $5, $6)"

	_, err := r.DB.Exec(ctx, query,
		userCexEchangeKeys.UserID,
		userCexEchangeKeys.CexName,
		userCexEchangeKeys.EncryptedApiKey,
		userCexEchangeKeys.EncryptedSecretKey,
		userCexEchangeKeys.EnrcyptedPassphrase,
		userCexEchangeKeys.EncryptedDek,
	)
	if err != nil {
		r.Log.Error("error with inserting cex exhcange api keys: ", "error", err)
		return err
	}

	return nil
}

func (r *RepositoryWallet) ActualizeSnapshot(
	ctx context.Context,
	userID uuid.UUID,
	accountType string,
	value float64,
	at time.Time,
) error {
	day := at.UTC().Truncate(24 * time.Hour)

	const query = `
		INSERT INTO snapshots(user_id, account_type, balance_value, at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id, at)
		DO UPDATE SET
			account_type  = EXCLUDED.account_type,
			balance_value = EXCLUDED.balance_value
	`
	if _, err := r.DB.Exec(ctx, query, userID, accountType, value, day); err != nil {
		r.Log.Error("error with actualizing snapshot", "error", err)
		return fmt.Errorf("actualize snapshot: %w", err)
	}
	return nil
}

func (r *RepositoryWallet) SelectExchangeApiKeys(ctx context.Context, userID uuid.UUID, cexName string) (dto.UserCexExchangeKeys, error) {
	query := "SELECT * FROM user_cex_exchange_keys WHERE user_id = $1 AND cex_name = $2"

	keys := dto.UserCexExchangeKeys{}

	if err := r.DB.QueryRow(ctx, query, userID, cexName).Scan(
		&keys.UserID,
		&keys.CexName,
		&keys.EncryptedApiKey,
		&keys.EncryptedSecretKey,
		&keys.EncryptedDek,
	); err != nil {
		r.Log.Error("error with selecting cex exchange api keys: ", "error", err)
		return dto.UserCexExchangeKeys{}, err
	}

	return keys, nil
}

func (r *RepositoryWallet) SelectSnapshotsByUserID(ctx context.Context, userID uuid.UUID) ([]dto.Snapshot, error) {
	query := "SELECT * FROM snapshots WHERE user_id = $1"

	var snapshots []dto.Snapshot

	rows, err := r.DB.Query(ctx, query, userID)
	if err != nil {
		r.Log.Error("error with selecting all exchanges by user id: ", "error", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		snap := dto.Snapshot{}
		if err := rows.Scan(
			&snap.UserID,
			&snap.AccountType,
			&snap.BalanceValue,
			&snap.At,
		); err != nil {
			r.Log.Error("error with scanning row to snapshot struct: ", "error", err)
			return nil, err
		}

		snapshots = append(snapshots, snap)
	}

	if err = rows.Err(); err != nil {
		r.Log.Error("error with getting rows from select all exchanges by user id: ", "error", err)
		return nil, err
	}

	return snapshots, nil
}

func (r *RepositoryWallet) UpdateExchangeApiKeys(ctx context.Context, userCexExchangeKeys dto.UserCexExchangeKeys) error {
	query := "UPDATE user_cex_exchange_keys SET encrypted_api_key = $3, encrypted_secret_key = $4, encrypted_dek = $5 WHERE user_id = $1 AND cex_name = $2"

	_, err := r.DB.Exec(ctx, query,
		userCexExchangeKeys.UserID,
		userCexExchangeKeys.CexName,
		userCexExchangeKeys.EncryptedApiKey,
		userCexExchangeKeys.EncryptedSecretKey,
		userCexExchangeKeys.EncryptedDek,
	)
	if err != nil {
		r.Log.Error("error with updating cex exchange api keys: ", "error", err)
		return err
	}

	return nil
}

func (r *RepositoryWallet) DeleteExchangeApiKeys(ctx context.Context, userID uuid.UUID, cexName string) error {
	query := "DELETE FROM user_cex_exchange_keys WHERE user_id = $1 AND cex_name = $2"

	_, err := r.DB.Exec(ctx, query, userID, cexName)
	if err != nil {
		r.Log.Error("error with deleting cex exchange api keys: ", "error", err)
		return err
	}

	return nil
}
