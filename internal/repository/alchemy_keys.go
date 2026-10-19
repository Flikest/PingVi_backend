package repository

import "context"

func (r *RepositoryWallet) AddAlchemyKey(ctx context.Context, apiKey string) error {
	query := `INSERT INTO alchemy_keys(api_key) VALUES ($1)`

	_, err := r.DB.Exec(ctx, query, apiKey)
	if err != nil {
		r.Log.Error("error with inserting api kety: ", "error", err)
		return err
	}

	return nil
}

func (r *RepositoryWallet) SelectActiveAlchemyKey(ctx context.Context) (int, string, error) {
	query := `
		SELECT api_key 
		FROM alchemy_keys 
		WHERE is_active = TRUE
		ORDER BY id ASC 
		LIMIT 1
	`

	var data struct {
		id     int
		apiKey string
	}

	if err := r.DB.QueryRow(ctx, query).Scan(&data); err != nil {
		r.Log.Error("error with selecting active api key: ", "error", err)
		return 0, "", err
	}

	return data.id, data.apiKey, nil
}

func (r *RepositoryWallet) UpdateTokenStatus(ctx context.Context, ID int) error {
	query := `
		UPDATE alchemy_keys
		SET is_active = NOT is_active
		WHERE id=$1
	`

	_, err := r.DB.Exec(ctx, query, ID)
	if err != nil {
		r.Log.Error("error with updating active status: ", "error", err)
		return err
	}

	return nil
}
