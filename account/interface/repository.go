package account

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/stevealexrs/Go-Libra/account/entity"
)

type sqlRepo struct {
	db *sql.DB
}

type UserRepo sqlRepo
type BusinessRepo sqlRepo

// The table name is in snake_case while the column name is in PascalCase
// TODO: Test this if you don't want unexpected thing to happen in database

func (r *UserRepo) Store(account *account.User) (int, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return 0, err
	}

	accStmt, err := tx.Prepare("INSERT INTO account VALUES(NULL, ?, ?, ?, ?);")
	if err != nil {
		tx.Rollback()
		return 0, err
	}
	defer accStmt.Close()

	res, err := accStmt.Exec(account.Username, account.PasswordHash, account.Email, account.UnverifiedEmail)
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	lastId, err := res.LastInsertId()
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	userAccStmt, err := tx.Prepare("INSERT INTO user VALUES(?, ?, ?)")
	if err != nil {
		tx.Rollback()
		return 0, err
	}
	defer userAccStmt.Close()

	_, err = userAccStmt.Exec(lastId, account.DisplayName, account.InvitationEmail)
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	return int(lastId), tx.Commit()
}

func (r *UserRepo) FetchById(id int) (*account.User, error) {
	query := "SELECT user.InvitationEmail, account.Username, user.DisplayName, account.PasswordHash, account.RecoveryEmail, account.UnverifiedRecoveryEmail " + 
			 "FROM user INNER JOIN account ON user.Id = account.Id WHERE user.Id = ? LIMIT 1;"

	stmt, err := r.db.Prepare(query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var passwordHash []byte
	var invitationEmail, username, displayName, email, unverifiedEmail string

	err = stmt.QueryRow(id).Scan(&invitationEmail, &username, &displayName, &passwordHash, &email, &unverifiedEmail)
	if err != nil {
		return nil, err
	}

	acc := &account.User{
		Base: account.Base{
			Id:              &id,
			Username:        username,
			PasswordHash:    passwordHash,
			Email:           email,
			UnverifiedEmail: unverifiedEmail,
		},
		InvitationEmail: invitationEmail,
		DisplayName:     displayName,
	}

	return acc, nil
}

func (r *UserRepo) FetchByUsername(name string) (*account.User, error) {
	query := "SELECT user.Id, user.InvitationEmail, user.DisplayName, account.PasswordHash, account.RecoveryEmail, account.UnverifiedRecoveryEmail " + 
			 "FROM user INNER JOIN account ON user.Id = account.Id WHERE account.Username = ? LIMIT 1;"

	stmt, err := r.db.Prepare(query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var id int
	var passwordHash []byte
	var invitationEmail, displayName, email, unverifiedEmail string

	err = stmt.QueryRow(name).Scan(&id, &invitationEmail, &displayName, &passwordHash, &email, &unverifiedEmail)
	if err != nil {
		return nil, err
	}

	acc := &account.User{
		Base: account.Base{
			Id:              &id,
			Username:        name,
			PasswordHash:    passwordHash,
			Email:           email,
			UnverifiedEmail: unverifiedEmail,
		},
		InvitationEmail: invitationEmail,
		DisplayName:     displayName,
	}

	return acc, nil
}

func (r *UserRepo) FetchByEmail(email string) ([]account.User, error) {
	if email == "" {
		return nil, errors.New("email cannot be empty")
	}

	query := "SELECT user.Id, user.InvitationEmail, account.Username, user.DisplayName, account.PasswordHash, account.UnverifiedRecoveryEmail " + 
			 "FROM user INNER JOIN account ON user.Id = account.Id WHERE account.RecoveryEmail = ?;"

	stmt, err := r.db.Prepare(query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(email)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accList []account.User
	var errList error

	for rows.Next() {
		var id int
		var passwordHash []byte
		var invitationEmail, username, displayName, unverifiedEmail string

		err = rows.Scan(&id, &invitationEmail, &username, &displayName, &passwordHash, &unverifiedEmail)
		if err != nil {
			errList = fmt.Errorf("%w; " + errList.Error(), err)
		}

		accList = append(accList, account.User{
			Base: account.Base{
				Id:              &id,
				Username:        username,
				PasswordHash:    passwordHash,
				Email:           email,
				UnverifiedEmail: unverifiedEmail,
			},
			InvitationEmail: invitationEmail,
			DisplayName:     displayName,
		})
	}
	if errList != nil {
		return accList, errList
	}

	return accList, rows.Err()
}

// Update Username, DisplayName, PasswordHash and Emails
func (r *UserRepo) Update(account *account.User) error {
	query := "UPDATE account, user " + 
			 "SET account.Username = ?, user.DisplayName = ?, account.PasswordHash = ?, account.RecoveryEmail = ?, account.UnverifiedRecoveryEmail = ? " +
			 "WHERE (user.Id = ? AND user.Id = account.Id);"

	stmt, err := r.db.Prepare(query)
	if err != nil {
		return err
	}

	_, err = stmt.Exec(account.Username, account.DisplayName, account.PasswordHash, account.Email, account.UnverifiedEmail, account.Id)
	return err
}

func (r *UserRepo) HasUsername(name string) (bool, error) {
	var username string
	// Fetch Random Data
	err := r.db.QueryRow("SELECT TOP 1 account.Username FROM account WHERE account.Username = ?;").Scan(&username)
	if err != nil && err != sql.ErrNoRows {
		return false, err
	}

	return err != sql.ErrNoRows, nil
}

func (r *UserRepo) HasInvitationEmail(email string) (bool, error) {
	var id int
	// Fetch Random Data
	err := r.db.QueryRow("SELECT TOP 1 user.Id FROM user WHERE user.InvitationEmail = ?;").Scan(&id)
	if err != nil && err != sql.ErrNoRows {
		return false, err
	}

	return err != sql.ErrNoRows, nil
}

func (r *BusinessRepo) Store(account *account.Business) (int, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return 0, err
	}

	accStmt, err := tx.Prepare("INSERT INTO account VALUES(NULL, ?, ?, ?, ?);")
	if err != nil {
		tx.Rollback()
		return 0, err
	}
	defer accStmt.Close()

	res, err := accStmt.Exec(account.Username, account.PasswordHash, account.Email, account.UnverifiedEmail)
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	lastId, err := res.LastInsertId()
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	businessStmt, err := tx.Prepare("INSERT INTO business VALUES(?, ?, ?)")
	if err != nil {
		tx.Rollback()
		return 0, err
	}
	defer businessStmt.Close()

	_, err = businessStmt.Exec(lastId, account.BusinessName.Name, account.BusinessName.Verified)
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	return int(lastId), tx.Commit()
}

func (r *BusinessRepo) StoreAsChild(account *account.Business, parentId int) (int, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return 0, err
	}

	accStmt, err := tx.Prepare("INSERT INTO account VALUES(NULL, ?, ?, ?, ?);")
	if err != nil {
		tx.Rollback()
		return 0, err
	}
	defer accStmt.Close()

	res, err := accStmt.Exec(account.Username, account.PasswordHash, account.Email, account.UnverifiedEmail)
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	lastId, err := res.LastInsertId()
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	businessStmt, err := tx.Prepare("INSERT INTO business VALUES(?, ?, ?)")
	if err != nil {
		tx.Rollback()
		return 0, err
	}
	defer businessStmt.Close()

	_, err = businessStmt.Exec(lastId, account.BusinessName.Name, account.BusinessName.Verified)
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	pcStmt, err := tx.Prepare("INSERT INTO business_parent_child VALUES(?, ?)")
	if err != nil {
		tx.Rollback()
		return 0, err
	}
	defer pcStmt.Close()

	_, err = pcStmt.Exec(parentId, lastId)
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	return int(lastId), tx.Commit()
}

func (r *BusinessRepo) StoreWithIdaccount(account *account.Business, identity *account.BusinessIdentity) (int, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return 0, err
	}

	accStmt, err := tx.Prepare("INSERT INTO account VALUES(NULL, ?, ?, ?, ?);")
	if err != nil {
		tx.Rollback()
		return 0, err
	}
	defer accStmt.Close()

	res, err := accStmt.Exec(account.Username, account.PasswordHash, account.Email, account.UnverifiedEmail)
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	lastId, err := res.LastInsertId()
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	businessStmt, err := tx.Prepare("INSERT INTO business VALUES(?, ?, ?)")
	if err != nil {
		tx.Rollback()
		return 0, err
	}
	defer businessStmt.Close()

	_, err = businessStmt.Exec(lastId, account.BusinessName.Name, account.BusinessName.Verified)
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	identityStmt, err := tx.Prepare("INSERT INTO business_identity VALUES(?, ?, ?, ?, ?, ?)")
	if err != nil {
		tx.Rollback()
		return 0, err
	}
	defer identityStmt.Close()

	_, err = identityStmt.Exec(lastId, identity.Name, identity.RegistrationNumber, identity.Address, identity.Document, identity.Verified)
	if err != nil {
		tx.Rollback()
		return 0, err
	}
	return int(lastId), tx.Commit()
}

func (r *BusinessRepo) FetchById(id int) (*account.Business, error) {
	query := "SELECT account.Username, business.DisplayName, business.DisplayNameVerified, account.PasswordHash, account.RecoveryEmail, account.UnverifiedRecoveryEmail " + 
			 "FROM user INNER JOIN account ON business.Id = account.Id WHERE business.Id = ? LIMIT 1;"

	stmt, err := r.db.Prepare(query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var passwordHash []byte
	var displayNameVerified bool
	var username, displayName, email, unverifiedEmail string

	err = stmt.QueryRow(id).Scan(&username, &displayName, &displayNameVerified, &passwordHash, &email, &unverifiedEmail)
	if err != nil {
		return nil, err
	}

	acc := &account.Business{
		Base: account.Base{
			Id:              &id,
			Username:        username,
			PasswordHash:    passwordHash,
			Email:           email,
			UnverifiedEmail: unverifiedEmail,
		},
		BusinessName: account.BusinessName{
			Name:     displayName,
			Verified: displayNameVerified,
		},
	}

	return acc, nil
}

func (r *BusinessRepo) FetchByUsername(name string) (*account.Business, error) {
	query := "SELECT business.Id, business.DisplayName, business.DisplayNameVerified, account.PasswordHash, account.RecoveryEmail, account.UnverifiedRecoveryEmail " + 
			 "FROM user INNER JOIN account ON business.Id = account.Id WHERE account.Username = ? LIMIT 1;"

	stmt, err := r.db.Prepare(query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var id int
	var passwordHash []byte
	var displayNameVerified bool
	var displayName, email, unverifiedEmail string

	err = stmt.QueryRow(name).Scan(&id, &displayName, &displayNameVerified, &passwordHash, &email, &unverifiedEmail)
	if err != nil {
		return nil, err
	}

	acc := &account.Business{
		Base: account.Base{
			Id:              &id,
			Username:        name,
			PasswordHash:    passwordHash,
			Email:           email,
			UnverifiedEmail: unverifiedEmail,
		},
		BusinessName: account.BusinessName{
			Name:     displayName,
			Verified: displayNameVerified,
		},
	}
	return acc, nil
}

func (r *BusinessRepo) FetchByEmail(email string) ([]account.Business, error) {
	if email == "" {
		return nil, errors.New("email cannot be empty")
	}

	query := "SELECT business.Id, account.Username, business.DisplayName, business.DisplayNameVerified, account.PasswordHash, account.UnverifiedRecoveryEmail " + 
			 "FROM user INNER JOIN account ON business.Id = account.Id WHERE account.RecoveryEmail = ? LIMIT 1;"

	stmt, err := r.db.Prepare(query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(email)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accList []account.Business
	var errList error

	for rows.Next() {
		var id int
		var passwordHash []byte
		var displayNameVerified bool
		var username, displayName, unverifiedEmail string

		err = rows.Scan(&id, &username, &displayName, &displayNameVerified, &passwordHash, &unverifiedEmail)
		if err != nil {
			errList = fmt.Errorf("%w; " + errList.Error(), err)
		}

		accList = append(accList, account.Business{
			Base: account.Base{
				Id:              &id,
				Username:        username,
				PasswordHash:    passwordHash,
				Email:           email,
				UnverifiedEmail: unverifiedEmail,
			},
			BusinessName: account.BusinessName{
				Name:     displayName,
				Verified: displayNameVerified,
			},
		})
	}
	if errList != nil {
		return accList, errList
	}

	return accList, rows.Err()
}

// Update Username, BusinessDisplayName, PasswordHash and Emails
func (r *BusinessRepo) Update(account *account.Business) error {
	query := "UPDATE account, business " + 
			 "SET account.Username = ?, business.DisplayName = ?, business.DisplayNameVerified, account.PasswordHash = ?, account.RecoveryEmail = ?, account.UnverifiedRecoveryEmail = ? " +
			 "WHERE (business.Id = ? AND business.Id = account.Id);"

	stmt, err := r.db.Prepare(query)
	if err != nil {
		return err
	}

	_, err = stmt.Exec(account.Username, account.BusinessName.Name, account.BusinessName.Verified, account.PasswordHash, account.Email, account.UnverifiedEmail, account.Id)
	return err
}

func (r *BusinessRepo) HasUsername(name string) (bool, error) {
	var username string
	// Fetch Random Data
	err := r.db.QueryRow("SELECT TOP 1 account.Username FROM account WHERE account.Username = ?;").Scan(&username)
	if err != nil && err != sql.ErrNoRows {
		return false, err
	}

	return err != sql.ErrNoRows, nil
}