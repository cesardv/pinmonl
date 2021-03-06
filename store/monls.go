package store

import (
	"context"
	"database/sql"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/pinmonl/pinmonl/database"
	"github.com/pinmonl/pinmonl/model"
)

type Monls struct {
	*Store
}

type MonlOpts struct {
	ListOpts
	URL           string
	FetchedBefore time.Time
}

func NewMonls(s *Store) *Monls {
	return &Monls{s}
}

func (m Monls) table() string {
	return "monls"
}

func (m *Monls) List(ctx context.Context, opts *MonlOpts) (model.MonlList, error) {
	if opts == nil {
		opts = &MonlOpts{}
	}

	qb := m.RunnableBuilder(ctx).
		Select(m.columns()...).From(m.table())
	qb = m.bindOpts(qb, opts)
	qb = addPagination(qb, opts)
	rows, err := qb.Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := make([]*model.Monl, 0)
	for rows.Next() {
		monl, err := m.scan(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, monl)
	}
	return list, nil
}

func (m *Monls) Count(ctx context.Context, opts *MonlOpts) (int64, error) {
	if opts == nil {
		opts = &MonlOpts{}
	}

	qb := m.RunnableBuilder(ctx).
		Select("count(*)").From(m.table())
	qb = m.bindOpts(qb, opts)
	row := qb.QueryRow()
	var count int64
	err := row.Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (m *Monls) Find(ctx context.Context, id string) (*model.Monl, error) {
	qb := m.RunnableBuilder(ctx).
		Select(m.columns()...).From(m.table()).
		Where("id = ?", id)
	row := qb.QueryRow()
	monl, err := m.scan(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return monl, nil
}

func (m *Monls) FindURL(ctx context.Context, url string) (*model.Monl, error) {
	qb := m.RunnableBuilder(ctx).
		Select(m.columns()...).From(m.table()).
		Where("url = ?", url)
	row := qb.QueryRow()
	monl, err := m.scan(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return monl, nil
}

func (m Monls) bindOpts(b squirrel.SelectBuilder, opts *MonlOpts) squirrel.SelectBuilder {
	if opts == nil {
		return b
	}

	if opts.URL != "" {
		b = b.Where("url = ?", opts.URL)
	}

	if !opts.FetchedBefore.IsZero() {
		b = b.Where("(fetched_at <= ? OR fetched_at IS NULL)", opts.FetchedBefore)
	}

	return b
}

func (m Monls) columns() []string {
	return []string{
		m.table() + ".id",
		m.table() + ".url",
		m.table() + ".fetched_at",
		m.table() + ".created_at",
		m.table() + ".updated_at",
	}
}

func (m Monls) scanColumns(monl *model.Monl) []interface{} {
	return []interface{}{
		&monl.ID,
		&monl.URL,
		&monl.FetchedAt,
		&monl.CreatedAt,
		&monl.UpdatedAt,
	}
}

func (m Monls) scan(row database.RowScanner) (*model.Monl, error) {
	var monl model.Monl
	err := row.Scan(m.scanColumns(&monl)...)
	if err != nil {
		return nil, err
	}
	return &monl, nil
}

func (m *Monls) Create(ctx context.Context, monl *model.Monl) error {
	monl2 := *monl
	monl2.ID = newID()
	monl2.CreatedAt = timestamp()
	monl2.UpdatedAt = timestamp()

	qb := m.RunnableBuilder(ctx).
		Insert(m.table()).
		Columns(
			"id",
			"url",
			"fetched_at",
			"created_at",
			"updated_at").
		Values(
			monl2.ID,
			monl2.URL,
			monl2.FetchedAt,
			monl2.CreatedAt,
			monl2.UpdatedAt)
	_, err := qb.Exec()
	if err != nil {
		return err
	}
	*monl = monl2
	return nil
}

func (m *Monls) Update(ctx context.Context, monl *model.Monl) error {
	monl2 := *monl
	monl2.UpdatedAt = timestamp()

	qb := m.RunnableBuilder(ctx).
		Update(m.table()).
		Set("url", monl2.URL).
		Set("fetched_at", monl2.FetchedAt).
		Set("updated_at", monl2.UpdatedAt).
		Where("id = ?", monl2.ID)
	_, err := qb.Exec()
	if err != nil {
		return err
	}
	*monl = monl2
	return nil
}

func (m *Monls) Delete(ctx context.Context, id string) (int64, error) {
	qb := m.RunnableBuilder(ctx).
		Delete(m.table()).
		Where("id = ?", id)
	res, err := qb.Exec()
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}
