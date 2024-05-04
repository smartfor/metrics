package storage

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/smartfor/metrics/internal/core"
	"github.com/smartfor/metrics/internal/server/utils"
	utils2 "github.com/smartfor/metrics/internal/utils"
)

type PostgresStorage struct {
	pool *pgxpool.Pool
}

func NewPostgresStorage(ctx context.Context, dsn string) (*PostgresStorage, error) {
	pool, err := utils2.Retry(func() (*pgxpool.Pool, error) {
		return pgxpool.New(ctx, dsn)
	}, nil)
	if err != nil {
		return nil, err
	}

	s := PostgresStorage{
		pool: pool,
	}

	if err := s.Initialize(); err != nil {
		return nil, err
	}

	return &s, nil
}

// Создает таблицы gauges и counters в базе данных
func (s *PostgresStorage) Initialize() error {
	return utils2.RetryVoid(s.initialize, nil)
}

func (s *PostgresStorage) SetBatch(ctx context.Context, batch core.BaseMetricStorage) error {
	return utils2.RetryVoid(func() error {
		return s.setBatch(ctx, batch)
	}, nil)
}

func (s *PostgresStorage) Set(metric core.MetricType, key string, value string) error {
	return utils2.RetryVoid(func() error {
		return s.set(metric, key, value)
	}, nil)
}

func (s *PostgresStorage) Get(metric core.MetricType, key string) (string, error) {
	switch metric {
	case core.Gauge:
		{
			v, err := utils2.Retry(func() (float64, error) {
				return s.getGauge(key)
			}, nil)
			if err != nil {
				return "", err
			}

			return utils.GaugeAsString(v), nil
		}
	case core.Counter:
		{
			d, err := utils2.Retry(func() (int64, error) {
				return s.getCounter(key)
			}, nil)
			if err != nil {
				return "", err
			}

			return utils.CounterAsString(d), nil
		}
	default:
		return "", core.ErrUnknownMetricType
	}
}

func (s *PostgresStorage) GetAll() (core.BaseMetricStorage, error) {
	return utils2.Retry(s.getAll, nil)
}

func (s *PostgresStorage) Close() error {
	s.pool.Close()
	return nil
}

func (s *PostgresStorage) Ping(ctx context.Context) error {
	return s.pool.Ping(ctx)
}

func (s *PostgresStorage) upsertGauge(tx pgx.Tx, key string, value float64) (pgconn.CommandTag, error) {
	return tx.Exec(
		context.TODO(),
		`INSERT INTO gauges (key, value)
			VALUES ($1, $2)
			ON CONFLICT (key)
			DO UPDATE SET value = EXCLUDED.value`,
		key, value,
	)
}

func (s *PostgresStorage) upsertCounter(tx pgx.Tx, key string, delta int64) (pgconn.CommandTag, error) {
	return tx.Exec(
		context.TODO(),
		`INSERT INTO counters (key, value)
		    	VALUES ($1, $2)
			ON CONFLICT (key)
			DO UPDATE SET value = counters.value + EXCLUDED.value`,
		key, delta,
	)
}

func (s *PostgresStorage) getGauge(key string) (float64, error) {
	var value float64
	err := s.pool.QueryRow(
		context.TODO(),
		`SELECT (value) FROM gauges WHERE key = $1 LIMIT 1`,
		key,
	).Scan(&value)

	if err != nil && errors.Is(err, pgx.ErrNoRows) {
		err = core.ErrNotFound
	}

	return value, err
}

func (s *PostgresStorage) getCounter(key string) (int64, error) {
	var delta int64
	err := s.pool.QueryRow(
		context.TODO(),
		`SELECT (value) FROM counters WHERE key = $1 LIMIT 1`,
		key,
	).Scan(&delta)

	if err != nil && errors.Is(err, pgx.ErrNoRows) {
		err = core.ErrNotFound
	}

	return delta, err
}

func (s *PostgresStorage) getAllGauges() (map[string]float64, error) {
	query, err := s.pool.Query(context.TODO(), `SELECT FROM gauges`)
	if err != nil {
		return nil, err
	}

	defer query.Close()

	rows := make(map[string]float64)
	for query.Next() {
		var id string
		var value float64
		if err := query.Scan(&id, &value); err != nil {
			return nil, err
		}
		rows[id] = value
	}

	return rows, nil
}

func (s *PostgresStorage) getAllCounters() (map[string]int64, error) {
	query, err := s.pool.Query(context.TODO(), `SELECT (key, value) FROM counters`)
	if err != nil {
		return nil, err
	}

	defer query.Close()

	rows := make(map[string]int64)
	for query.Next() {
		var id string
		var value int64
		if err := query.Scan(&id, &value); err != nil {
			return nil, err
		}
		rows[id] = value
	}

	return rows, nil
}

func (s *PostgresStorage) getAll() (core.BaseMetricStorage, error) {
	gauges, err := s.getAllGauges()
	if err != nil {
		return core.BaseMetricStorage{}, err
	}

	counters, err := s.getAllCounters()
	if err != nil {
		return core.BaseMetricStorage{}, err
	}

	return core.NewBaseMetricStorageWithValues(gauges, counters), nil
}

func (s *PostgresStorage) initialize() error {
	_, err := s.pool.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS gauges (
			key VARCHAR(255) PRIMARY KEY,
			value DOUBLE PRECISION
		);
	`)
	if err != nil {
		return err
	}

	_, err = s.pool.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS counters (
			key VARCHAR(255) PRIMARY KEY,
			value INT8
		);
	`)
	if err != nil {
		return err
	}
	return nil
}

func (s *PostgresStorage) set(metric core.MetricType, key string, value string) error {
	ctx := context.TODO()
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	switch metric {
	case core.Gauge:
		{
			val, err := utils.GaugeFromString(value)
			if err != nil {
				return core.ErrBadMetricValue
			}

			_, err = s.upsertGauge(tx, key, val)
			if err != nil {
				return err
			}

			err = tx.Commit(ctx)
			if err != nil {
				return err
			}
		}

	case core.Counter:
		{
			delta, err := utils.CounterFromString(value)
			if err != nil {
				return core.ErrBadMetricValue
			}

			_, err = s.upsertCounter(tx, key, delta)
			if err != nil {
				return err
			}

			err = tx.Commit(ctx)
			if err != nil {
				return err
			}
		}

	default:
		return core.ErrUnknownMetricType
	}

	return nil
}

func (s *PostgresStorage) setBatch(ctx context.Context, batch core.BaseMetricStorage) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for k, v := range batch.Gauges() {
		if _, err := s.upsertGauge(tx, k, v); err != nil {
			return err
		}
	}

	for k, v := range batch.Counters() {
		if _, err := s.upsertCounter(tx, k, v); err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}
