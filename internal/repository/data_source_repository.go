package repository

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"

	"github.com/medivh13/synaptica-api/internal/model"
)

type DataSourceRepository interface {
	FindByType(ctx context.Context, sourceType string) (*model.DataSource, error)
}

type dataSourceRepository struct {
	db *sqlx.DB
}

func NewDataSourceRepository(db *sqlx.DB) DataSourceRepository {
	return &dataSourceRepository{db: db}
}

func (r *dataSourceRepository) FindByType(ctx context.Context, sourceType string) (*model.DataSource, error) {
	const query = `
SELECT id, name, type, base_url, created_at
FROM data_sources
WHERE type = $1
LIMIT 1;
`

	var source model.DataSource
	if err := r.db.GetContext(ctx, &source, query, sourceType); err != nil {
		return nil, fmt.Errorf("find data source by type %q: %w", sourceType, err)
	}

	return &source, nil
}
