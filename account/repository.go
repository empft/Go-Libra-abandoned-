package account

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"log"
	"strings"

	"github.com/stevealexrs/Go-Libra/database/object"
	"github.com/stevealexrs/Go-Libra/database/sqltype"
)

type UserRepo struct {
	DB *sql.DB
}
type BusinessRepo struct {
	DB *sql.DB
	objStore object.Store
}

// The table name is in snake_case while the column name is in PascalCase
// TODO: Test this if you don't want unexpected thing to happen in database

func (r *UserRepo) Store(ctx context.Context, account *User) (int, error) {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}

	accStmt, err := tx.PrepareContext(ctx, "INSERT INTO account VALUES(NULL, ?, ?, ?, ?, ?);")
	if err != nil {
		tx.Rollback()
		return 0, err
	}
	defer accStmt.Close()

	res, err := accStmt.ExecContext(ctx, account.Username, account.PasswordHash, account.Email, account.UnverifiedEmail, sqltype.MyBool(account.Deleted))
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	lastId, err := res.LastInsertId()
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	userAccStmt, err := tx.PrepareContext(ctx, "INSERT INTO user VALUES(?, ?, ?)")
	if err != nil {
		tx.Rollback()
		return 0, err
	}
	defer userAccStmt.Close()

	_, err = userAccStmt.ExecContext(ctx, lastId, account.DisplayName, account.InvitationEmail)
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	return int(lastId), tx.Commit()
}

func (r *UserRepo) FetchById(ctx context.Context, id int) (*User, error) {
	query := "SELECT user.InvitationEmail, account.Username, user.DisplayName, account.PasswordHash, " +
			 "account.RecoveryEmail, account.UnverifiedRecoveryEmail, account.Deleted " + 
			 "FROM user INNER JOIN account ON user.Id = account.Id WHERE user.Id = ? LIMIT 1;"

	stmt, err := r.DB.PrepareContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var passwordHash []byte
	var invitationEmail, username, displayName, email, unverifiedEmail string
	var deleted sqltype.MyBool

	err = stmt.QueryRowContext(ctx, id).Scan(&invitationEmail, &username, &displayName, &passwordHash, &email, &unverifiedEmail, &deleted)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errDoesNotExist
	} else if err != nil {
		return nil, err
	}

	acc := &User{
		base: base{
			Id:              &id,
			Username:        username,
			PasswordHash:    passwordHash,
			Email:           email,
			UnverifiedEmail: unverifiedEmail,
			Deleted: bool(deleted),
		},
		InvitationEmail: invitationEmail,
		DisplayName:     displayName,
	}

	return acc, nil
}

func (r *UserRepo) FetchByUsername(ctx context.Context, name string) (*User, error) {
	query := "SELECT user.Id, user.InvitationEmail, user.DisplayName, account.PasswordHash, " +
			 "account.RecoveryEmail, account.UnverifiedRecoveryEmail, account.Deleted " + 
			 "FROM user INNER JOIN account ON user.Id = account.Id WHERE account.Username = ? LIMIT 1;"

	stmt, err := r.DB.PrepareContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var id int
	var passwordHash []byte
	var invitationEmail, displayName, email, unverifiedEmail string
	var deleted sqltype.MyBool

	err = stmt.QueryRowContext(ctx, name).Scan(&id, &invitationEmail, &displayName, &passwordHash, &email, &unverifiedEmail, &deleted)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errDoesNotExist
	} else if err != nil {
		return nil, err
	}

	acc := &User{
		base: base{
			Id:              &id,
			Username:        name,
			PasswordHash:    passwordHash,
			Email:           email,
			UnverifiedEmail: unverifiedEmail,
			Deleted: bool(deleted),
		},
		InvitationEmail: invitationEmail,
		DisplayName:     displayName,
	}

	return acc, nil
}

func (r *UserRepo) FetchByEmail(ctx context.Context, email string) ([]User, error) {
	if email == "" {
		return nil, errors.New("email cannot be empty")
	}

	query := "SELECT user.Id, user.InvitationEmail, account.Username, user.DisplayName, account.PasswordHash, " +
			 "account.UnverifiedRecoveryEmail, account.Deleted " + 
			 "FROM user INNER JOIN account ON user.Id = account.Id WHERE account.RecoveryEmail = ?;"

	stmt, err := r.DB.PrepareContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.QueryContext(ctx, email)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accList []User

	for rows.Next() {
		var id int
		var passwordHash []byte
		var invitationEmail, username, displayName, unverifiedEmail string
		var deleted sqltype.MyBool

		err = rows.Scan(&id, &invitationEmail, &username, &displayName, &passwordHash, &unverifiedEmail, &deleted)
		if err != nil {
			return nil, err
		}

		accList = append(accList, User{
			base: base{
				Id:              &id,
				Username:        username,
				PasswordHash:    passwordHash,
				Email:           email,
				UnverifiedEmail: unverifiedEmail,
				Deleted: bool(deleted),
			},
			InvitationEmail: invitationEmail,
			DisplayName:     displayName,
		})
	}
	return accList, rows.Err()
}

func (r *UserRepo) Update(ctx context.Context, account *User) error {
	query := "UPDATE account, user " + 
			 "SET account.Username = ?, user.DisplayName = ?, account.PasswordHash = ?, " +
			 "account.RecoveryEmail = ?, account.UnverifiedRecoveryEmail = ?, account.Deleted = ? " +
			 "WHERE (user.Id = ? AND user.Id = account.Id);"

	stmt, err := r.DB.PrepareContext(ctx, query)
	if err != nil {
		return err
	}

	_, err = stmt.ExecContext(
		ctx,
		account.Username,
		account.DisplayName,
		account.PasswordHash,
		account.Email,
		account.UnverifiedEmail,
		sqltype.MyBool(account.Deleted),
		account.Id,
	)
	return err
}

func (r *UserRepo) HasUsername(ctx context.Context, name string) (bool, error) {
	var username string
	// Fetch Random Data
	err := r.DB.QueryRowContext(ctx, "SELECT account.Username FROM account WHERE account.Username = ? LIMIT 1;").Scan(&username)
	if err != nil && err != sql.ErrNoRows {
		return false, err
	}

	return err != sql.ErrNoRows, nil
}

func (r *UserRepo) HasInvitationEmail(ctx context.Context, email string) (bool, error) {
	var id int
	// Fetch Random Data
	err := r.DB.QueryRowContext(ctx, "SELECT user.Id FROM user WHERE user.InvitationEmail = ? LIMIT 1;").Scan(&id)
	if err != nil && err != sql.ErrNoRows {
		return false, err
	}

	return err != sql.ErrNoRows, nil
}

const fileIdSeparator string = ","

func serializeFileId(fids []string) string {
	return strings.Join(fids, fileIdSeparator)
}

func unserializeFileId(fids string) []string {
	return strings.Split(fids, fileIdSeparator)
}

// returns file id after storing the files
func (r *BusinessRepo) AddFilesToObjectStore(ctx context.Context, documents ...[]byte) ([]string, error) {
	fids := []string{}

	// add documents to object store
	for _, v := range documents {
		fid, err := r.objStore.Set(ctx, bytes.NewReader(v))
		if err != nil {
			// Delete added objects
			for _, v := range fids {
				err = r.objStore.Delete(ctx, v)
				if err != nil {
					// ALERT LOG
					log.Printf("fail to delete object with id %s: %s/n", v, err)
				}
			}
			return nil, err
		}
		fids = append(fids, fid)
	}
	return fids, nil
}

func (r *BusinessRepo) Store(ctx context.Context, account *Business, documents [][]byte) (int, error) {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}

	accStmt, err := tx.PrepareContext(ctx, "INSERT INTO account VALUES(NULL, ?, ?, ?, ?, ?);")
	if err != nil {
		tx.Rollback()
		return 0, err
	}
	defer accStmt.Close()

	res, err := accStmt.ExecContext(
		ctx,
		account.Username,
		account.PasswordHash,
		account.Email,
		account.UnverifiedEmail,
		sqltype.MyBool(account.Deleted),
	)
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	lastId, err := res.LastInsertId()
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	businessStmt, err := tx.PrepareContext(ctx, "INSERT INTO business VALUES(?, ?, ?)")
	if err != nil {
		tx.Rollback()
		return 0, err
	}
	defer businessStmt.Close()

	_, err = businessStmt.ExecContext(ctx, lastId, account.BusinessName.DisplayName, account.BusinessName.Verified)
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	pcStmt, err := tx.PrepareContext(ctx, "INSERT INTO business_parent_child VALUES(?, ?)")
	if err != nil {
		tx.Rollback()
		return 0, err
	}
	defer pcStmt.Close()

	_, err = pcStmt.ExecContext(ctx, *account.ChildOf, lastId)
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	identity := account.BusinessIdentity
	if identity.Name != "" ||
	   identity.Address != "" ||
	   identity.RegistrationNumber != "" ||
	   identity.Verified ||
	   documents != nil {

		fids := []string{}
		if documents != nil {
			fids, err = r.AddFilesToObjectStore(ctx, documents...)
			if err != nil {
				tx.Rollback()
				return 0, err
			}
		}
		// if any error happens down there, the files become orphan objects

		identityStmt, err := tx.PrepareContext(ctx, "INSERT INTO business_identity VALUES(?, ?, ?, ?, ?, ?)")
		if err != nil {
			tx.Rollback()
			// ALERT log
			log.Printf("database transaction failed, files %s become orphan objects: %s", fids, err)
			return 0, err
		}
		defer identityStmt.Close()

		serialized := serializeFileId(fids)
		identity := account.BusinessIdentity
		_, err = identityStmt.ExecContext(ctx, lastId, identity.Name, identity.RegistrationNumber, identity.Address, serialized, identity.Verified)
		if err != nil {
			tx.Rollback()
			// ALERT log
			log.Printf("database transaction failed, files %s become orphan objects: %s", fids, err)
			return 0, err
		}
	}
	return int(lastId), tx.Commit()
}

func (r *BusinessRepo) FetchById(ctx context.Context, id int) (*Business, error) {
	query := "SELECT acc.Username, b.DisplayName, b.DisplayNameVerified, acc.PasswordHash, " +
			 "acc.RecoveryEmail, acc.UnverifiedRecoveryEmail, acc.Deleted, " +
			 "bi.BusinessOfficialName, bi.BusinessRegistrationNumber, bi.BusinessAddress, bi.Documents, bi.Verified, " +
			 "COALESCE(child.ParentId, -1) " +
			 "FROM account AS acc " +
			 "INNER JOIN business AS b ON b.Id = acc.Id " +
			 "INNER JOIN business_identity AS bi ON bi.Id = acc.Id " +
			 "LEFT JOIN business_parent_child AS child ON child.ChildId = acc.Id WHERE acc.Id = ? LIMIT 1;"

	stmt, err := r.DB.PrepareContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var passwordHash []byte
	var displayNameVerified bool
	var username, displayName, email, unverifiedEmail, businessName, businessRegNum, businessAddr, businessDocuments string
	var deleted, businessVerified sqltype.MyBool
	var parent int

	err = stmt.QueryRowContext(ctx, id).Scan(
		&username,
		&displayName,
		&displayNameVerified,
		&passwordHash,
		&email,
		&unverifiedEmail,
		&deleted,
		&businessName,
		&businessRegNum,
		&businessAddr,
		&businessDocuments,
		&businessVerified,
		&parent,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errDoesNotExist
	} else if err != nil {
		return nil, err
	}

	fidURL, err := r.objStore.FormatURL(unserializeFileId(businessDocuments)...)
	if err != nil {
		return nil, err
	}

	var child *int
	if parent == -1 {
		child = nil
	} else {
		child = &parent
	}

	acc := &Business{
		base: base{
			Id:              &id,
			Username:        username,
			PasswordHash:    passwordHash,
			Email:           email,
			UnverifiedEmail: unverifiedEmail,
			Deleted: 		 bool(deleted),
		},
		BusinessName: BusinessName{
			DisplayName: displayName,
			Verified:	 displayNameVerified,
		},
		BusinessIdentity: BusinessIdentity{
			Name:               businessName,
			RegistrationNumber: businessRegNum,
			Address:            businessAddr,
			Documents:          fidURL,
			Verified:           bool(businessVerified),
		},
		ChildOf: child,
	}

	return acc, nil
}

func (r *BusinessRepo) FetchByUsername(ctx context.Context, name string) (*Business, error) {
	query := "SELECT acc.Id, b.DisplayName, b.DisplayNameVerified, acc.PasswordHash, " +
			 "acc.RecoveryEmail, acc.UnverifiedRecoveryEmail, acc.Deleted, " +
			 "bi.BusinessOfficialName, bi.BusinessRegistrationNumber, bi.BusinessAddress, bi.Documents, bi.Verified, " +
			 "COALESCE(child.ParentId, -1) " +
			 "FROM account AS acc " +
			 "INNER JOIN business AS b ON b.Id = acc.Id " +
			 "INNER JOIN business_identity AS bi ON bi.Id = acc.Id " +
			 "LEFT JOIN business_parent_child AS child ON child.ChildId = acc.Id WHERE acc.Username = ? LIMIT 1;"

	stmt, err := r.DB.PrepareContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var id, parent int
	var passwordHash []byte
	var displayNameVerified bool
	var displayName, email, unverifiedEmail, businessName, businessRegNum, businessAddr, businessDocuments string
	var deleted, businessVerified sqltype.MyBool

	err = stmt.QueryRowContext(ctx, name).Scan(
		&id,
		&displayName,
		&displayNameVerified,
		&passwordHash,
		&email,
		&unverifiedEmail,
		&deleted,
		&businessName,
		&businessRegNum,
		&businessAddr,
		&businessDocuments,
		&businessVerified,
		&parent,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errDoesNotExist
	} else if err != nil {
		return nil, err
	}

	fidURL, err := r.objStore.FormatURL(unserializeFileId(businessDocuments)...)
	if err != nil {
		return nil, err
	}

	var child *int
	if parent == -1 {
		child = nil
	} else {
		child = &parent
	}

	acc := &Business{
		base: base{
			Id:              &id,
			Username:        name,
			PasswordHash:    passwordHash,
			Email:           email,
			UnverifiedEmail: unverifiedEmail,
			Deleted: 		 bool(deleted),
		},
		BusinessName: BusinessName{
			DisplayName: displayName,
			Verified:	 displayNameVerified,
		},
		BusinessIdentity: BusinessIdentity{
			Name:               businessName,
			RegistrationNumber: businessRegNum,
			Address:            businessAddr,
			Documents:          fidURL,
			Verified:           bool(businessVerified),
		},
		ChildOf: child,
	}
	return acc, nil
}

func (r *BusinessRepo) FetchByEmail(ctx context.Context, email string) ([]Business, error) {
	if email == "" {
		return nil, errors.New("email cannot be empty")
	}
	query := "SELECT acc.Id, acc.Username, b.DisplayName, b.DisplayNameVerified, acc.PasswordHash, " +
			 "acc.UnverifiedRecoveryEmail, acc.Deleted, " +
			 "bi.BusinessOfficialName, bi.BusinessRegistrationNumber, bi.BusinessAddress, bi.Documents, bi.Verified, " +
			 "COALESCE(child.ParentId, -1) " +
			 "FROM account AS acc " +
			 "INNER JOIN business AS b ON b.Id = acc.Id " +
			 "INNER JOIN business_identity AS bi ON bi.Id = acc.Id " +
			 "LEFT JOIN business_parent_child AS child ON child.ChildId = acc.Id WHERE acc.RecoveryEmail = ?;"

	stmt, err := r.DB.PrepareContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.QueryContext(ctx, email)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accList []Business

	for rows.Next() {
		var id, parent int
		var passwordHash []byte
		var displayNameVerified bool
		var username, displayName, unverifiedEmail, businessName, businessRegNum, businessAddr, businessDocuments string
		var deleted, businessVerified sqltype.MyBool

		err = rows.Scan(
			&id,
			&username,
			&displayName,
			&displayNameVerified,
			&passwordHash,
			&unverifiedEmail,
			&deleted,
			&businessName,
			&businessRegNum,
			&businessAddr,
			&businessDocuments,
			&businessVerified,
			&parent,
		)
		if err != nil {
			return nil, err
		}

		fidURL, err := r.objStore.FormatURL(unserializeFileId(businessDocuments)...)
		if err != nil {
			return nil, err
		}

		var child *int
		if parent == -1 {
			child = nil
		} else {
			child = &parent
		}

		accList = append(accList, Business{
			base: base{
				Id:              &id,
				Username:        username,
				PasswordHash:    passwordHash,
				Email:           email,
				UnverifiedEmail: unverifiedEmail,
				Deleted: 		 bool(deleted),
			},
			BusinessName: BusinessName{
				DisplayName: displayName,
				Verified:	 displayNameVerified,
			},
			BusinessIdentity: BusinessIdentity{
				Name:               businessName,
				RegistrationNumber: businessRegNum,
				Address:            businessAddr,
				Documents:          fidURL,
				Verified:           bool(businessVerified),
			},
			ChildOf: child,
		})
	}
	return accList, rows.Err()
}

func (r *BusinessRepo) Update(ctx context.Context, account *Business, documents [][]byte) error {
	var old string
	err := r.DB.QueryRowContext(ctx, "SELECT bi.Documents FROM business_identity AS bi WHERE bi.AccountId = ? LIMIT 1;", account.Id).Scan(&old)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrAccountNotExist(ctx)
		}
		return err
	}

	newFID, err := r.objStore.FormatFID(account.Documents...)
	if err != nil {
		return err
	}

	newFIDMap := make(map[string]struct{})
	for _, v := range newFID {
        newFIDMap[v] = struct{}{}
    }

	// fid that is in old list but not in new list
	removeFID := make([]string, 0)
	for _, v := range unserializeFileId(old) {
		if _, found := newFIDMap[v]; !found {
			removeFID = append(removeFID, v)
		} 
	}

	// delete objects, does not mind error
	err = r.objStore.Delete(ctx, removeFID...)
	if err != nil {
		// ALERT log
		log.Printf("fail to delete objects %v: %s", removeFID, err)
	}

	addFID, err := r.AddFilesToObjectStore(ctx, documents...)
	if err != nil {
		return err
	}

	newFID = append(newFID, addFID...)

	query := "UPDATE account AS acc, business AS b, business_identity AS bi" + 
			 "SET acc.Username = ?, b.DisplayName = ?, b.DisplayNameVerified, acc.PasswordHash = ?, acc.RecoveryEmail = ?, acc.UnverifiedRecoveryEmail = ?, " +
			 "bi.BusinessOfficialName = ?, bi.BusinessRegistrationNumber = ?, bi.BusinessAddress = ?, bi.Documents = ?, bi.Verified = ? " +
			 "WHERE (b.Id = ? AND b.Id = acc.Id AND b.Id = bi.AccountId);"

	stmt, err := r.DB.PrepareContext(ctx, query)
	if err != nil {
		return err
	}

	_, err = stmt.ExecContext(
		ctx,
		account.Username,
		account.BusinessName.DisplayName,
		account.BusinessName.Verified,
		account.PasswordHash,
		account.Email,
		account.UnverifiedEmail,
		account.BusinessName,
		account.RegistrationNumber,
		account.Address,
		serializeFileId(newFID),
		account.BusinessIdentity.Verified,
		account.Id,
	)
	return err
}

func (r *BusinessRepo) HasUsername(ctx context.Context, name string) (bool, error) {
	var username string
	// Fetch Random Data
	err := r.DB.QueryRowContext(ctx, "SELECT account.Username FROM account WHERE account.Username = ? LIMIT 1;", name).Scan(&username)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return false, err
	}

	return !errors.Is(err, sql.ErrNoRows), nil
}