package store

import (
	"context"
	"database/sql"

	"github.com/pinmonl/pinmonl/database"
	"github.com/pinmonl/pinmonl/model"
)

// UserOpts defines the parameters for user filtering.
type UserOpts struct {
	ListOpts
	Login string
}

// UserStore defines the services of user.
type UserStore interface {
	List(context.Context, *UserOpts) ([]model.User, error)
	Find(context.Context, *model.User) error
	FindLogin(context.Context, *model.User) error
	Create(context.Context, *model.User) error
	Update(context.Context, *model.User) error
	Delete(context.Context, *model.User) error
}

// NewUserStore creates user store.
func NewUserStore(s Store) UserStore {
	return &dbUserStore{s}
}

type dbUserStore struct {
	Store
}

// List retrieves users by the filter parameters.
func (s *dbUserStore) List(ctx context.Context, opts *UserOpts) ([]model.User, error) {
	e := s.Queryer(ctx)
	br, args := bindUserOpts(opts)
	rows, err := e.NamedQuery(br.String(), args)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ms []model.User
	for rows.Next() {
		var m model.User
		err = rows.StructScan(&m)
		if err != nil {
			return nil, err
		}
		ms = append(ms, m)
	}
	return ms, nil
}

// Find retrieves user by id.
func (s *dbUserStore) Find(ctx context.Context, m *model.User) error {
	e := s.Queryer(ctx)
	br, _ := bindUserOpts(nil)
	br.Where = []string{"id = :id"}
	br.Limit = 1
	rows, err := e.NamedQuery(br.String(), m)
	if err != nil {
		return err
	}
	defer rows.Close()

	if !rows.Next() {
		return sql.ErrNoRows
	}
	var m2 model.User
	err = rows.StructScan(&m2)
	if err != nil {
		return err
	}
	*m = m2
	return nil
}

// FindLogin retrieves user by login.
func (s *dbUserStore) FindLogin(ctx context.Context, m *model.User) error {
	e := s.Queryer(ctx)
	br, _ := bindUserOpts(nil)
	br.Where = []string{"login = :login"}
	br.Limit = 1
	rows, err := e.NamedQuery(br.String(), m)
	if err != nil {
		return err
	}
	defer rows.Close()

	if !rows.Next() {
		return sql.ErrNoRows
	}
	var m2 model.User
	err = rows.StructScan(&m2)
	if err != nil {
		return err
	}
	*m = m2
	return nil
}

// Create inserts the fields of user with generated id.
func (s *dbUserStore) Create(ctx context.Context, m *model.User) error {
	m2 := *m
	m2.ID = newUID()
	m2.CreatedAt = timestamp()
	e := s.Execer(ctx)
	stmt := database.InsertBuilder{
		Into: userTB,
		Fields: map[string]interface{}{
			"id":         nil,
			"login":      nil,
			"password":   nil,
			"name":       nil,
			"image_id":   nil,
			"role":       nil,
			"hash":       nil,
			"created_at": nil,
			"last_log":   nil,
		},
	}.String()
	_, err := e.NamedExec(stmt, m2)
	if err != nil {
		return err
	}
	*m = m2
	return nil
}

// Update updates the fields of user by id.
func (s *dbUserStore) Update(ctx context.Context, m *model.User) error {
	m2 := *m
	m2.UpdatedAt = timestamp()
	e := s.Execer(ctx)
	stmt := database.UpdateBuilder{
		From: userTB,
		Fields: map[string]interface{}{
			"login":      nil,
			"password":   nil,
			"name":       nil,
			"image_id":   nil,
			"role":       nil,
			"hash":       nil,
			"updated_at": nil,
			"last_log":   nil,
		},
		Where: []string{"id = :id"},
	}.String()
	_, err := e.NamedExec(stmt, m2)
	if err != nil {
		return err
	}
	*m = m2
	return nil
}

// Delete removes user by id.
func (s *dbUserStore) Delete(ctx context.Context, m *model.User) error {
	e := s.Execer(ctx)
	stmt := database.DeleteBuilder{
		From:  userTB,
		Where: []string{"id = :id"},
	}.String()
	_, err := e.NamedExec(stmt, m)
	return err
}

func bindUserOpts(opts *UserOpts) (database.SelectBuilder, map[string]interface{}) {
	br := database.SelectBuilder{
		From: userTB,
		Columns: database.NamespacedColumn(
			[]string{"id", "login", "password", "name", "image_id", "role", "hash", "created_at", "updated_at", "last_log"},
			userTB,
		),
	}
	if opts == nil {
		return br, nil
	}

	br = appendListOpts(br, opts.ListOpts)
	args := make(map[string]interface{})

	if opts.Login != "" {
		br.Where = append(br.Where, "login = :login")
		args["login"] = opts.Login
	}

	return br, args
}
