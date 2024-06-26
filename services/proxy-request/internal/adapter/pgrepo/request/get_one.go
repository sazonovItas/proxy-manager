package pgrequest

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/sazonovItas/proxy-manager/services/proxy-request/internal/adapter"
	"github.com/sazonovItas/proxy-manager/services/proxy-request/internal/entity"
)

func (rr *requestRepository) Request(ctx context.Context, id uuid.UUID) (*entity.Request, error) {
	const op = "adapter.pgrepo.request.Request"

	const query = "SELECT * FROM %s WHERE id=$1"

	stmt, err := rr.db.PreparexContext(ctx, rr.table(query))
	if err != nil {
		return nil, fmt.Errorf("%s: failed prepare statement: %w", op, err)
	}
	defer stmt.Close()

	var request entity.Request
	err = stmt.GetContext(ctx, &request, id)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, adapter.ErrRequestNotFound
		default:
			return nil, fmt.Errorf("%s: failed get request: %w", op, err)
		}
	}

	return &request, nil
}
