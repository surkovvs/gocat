package sqlcbatches

import (
	"errors"

	pgx "github.com/jackc/pgx/v5"
	"github.com/samber/lo"
)

type queryerOne[R any] interface {
	QueryRow(f func(int, R, error))
	Close() error
}

// 'ErrBatchAlreadyClosed' must be passed from sqlc generated pkg
func GetBatchOne[R, T any](q queryerOne[R], conv func(R) T, preSize int, ErrBatchAlreadyClosed error) ([]T, error) {
	ent := make([]T, 0, preSize)
	var queryErr error
	q.QueryRow(func(i int, dto R, err error) {
		switch {
		// если querier уже закрыт, безымянная функция все равно будет вызываться,
		// чтобы не затереть полученную ошибку - скипаем
		case errors.Is(err, ErrBatchAlreadyClosed):
		// если по аргументам запроса ничего не найдено - скипаем
		case errors.Is(err, pgx.ErrNoRows):
		// при неожиданной ошибке, предполагаем закрытие вопрошатора
		case err != nil:
			queryErr = err
			q.Close()
		default:
			ent = append(ent, conv(dto))
		}
	})
	if queryErr != nil {
		return nil, queryErr
	}

	return ent, nil
}

type queryerMany[R any] interface {
	Query(f func(int, []R, error))
	Close() error
}

// 'ErrBatchAlreadyClosed' must be passed from sqlc generated pkg
func GetBatchMany[R, T any](q queryerMany[R], conv func(R) T, preSize int, ErrBatchAlreadyClosed error) ([]T, error) {
	ent := make([]T, 0, preSize)
	var queryErr error
	q.Query(func(i int, dto []R, err error) {
		switch {
		// если querier уже закрыт, безымянная функция все равно будет вызываться,
		// чтобы не затереть полученную ошибку - скипаем
		case errors.Is(err, ErrBatchAlreadyClosed):
		// если по аргументам запроса ничего не найдено - скипаем
		case errors.Is(err, pgx.ErrNoRows):
		// при неожиданной ошибке, предполагаем закрытие вопрошатора
		case err != nil:
			queryErr = err
			q.Close()
		default:
			ent = append(ent, lo.Map(dto, func(p R, _ int) T {
				return conv(p)
			})...)
		}
	})
	if queryErr != nil {
		return nil, queryErr
	}

	return ent, nil
}
