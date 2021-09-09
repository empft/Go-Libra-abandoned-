package account_test

import (
	"bytes"
	"context"
	"database/sql"
	"image"
	"image/color"
	"image/png"
	"io"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stevealexrs/Go-Libra/account"
	"github.com/stevealexrs/Go-Libra/database/object"
)

var userRepo *account.UserRepo
var businessRepo *account.BusinessRepo
var objectMock = object.NewMock("")

// connect to database and reset database
func init() {
	tableName := "libra"
	sqlDB, err := sql.Open("mysql", "root:pass@tcp(0.0.0.0:4444)/" + tableName)
	if err != nil {
		panic(err)
	}

	userRepo = &account.UserRepo{
		DB: sqlDB,
	}
	businessRepo = &account.BusinessRepo{
		DB: sqlDB,
		ObjStore: objectMock,
	}

	tx, err := sqlDB.Begin()
	if err != nil {
		panic(err)
	}

	_, err = tx.Exec("SET FOREIGN_KEY_CHECKS=0;")
	if err != nil {
		panic(err)
	}

	rows, err := tx.Query("SELECT Concat('TRUNCATE TABLE ',table_schema,'.',TABLE_NAME, ';') FROM INFORMATION_SCHEMA.TABLES where table_schema in ('" + tableName + "');")
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	var queries []string
	for rows.Next() {
		var query string
		err := rows.Scan(&query)
		if err != nil {
			panic(err)
		}
		queries = append(queries, query)
	}
	if rows.Err() != nil {
		panic(rows.Err())
	}

	for _, v := range queries {
		_, err = tx.Exec(v)
		if err != nil {
			panic(err)
		}
	}

	_, err = tx.Exec("SET FOREIGN_KEY_CHECKS=1;")
	if err != nil {
		panic(err)
	}

	if err = tx.Commit(); err != nil {
		panic(err)
	}
}

func cmpUser(l, r account.User) bool {
	if equal := cmp.Equal(l, r, cmpopts.IgnoreFields(account.User{}, "Id", "PasswordHash")); !equal {
		return false
	}
	if l.Id != nil && r.Id != nil && *l.Id != *r.Id {
		return false
	}
	if !bytes.Equal(l.PasswordHash, r.PasswordHash)  {
		return false
	}
	return true
}

func cmpBusiness(l, r account.Business) bool {
	if equal := cmp.Equal(l, r, cmpopts.IgnoreFields(account.Business{}, "Id", "PasswordHash", "ChildOf", "BusinessIdentity.Documents")); !equal {
		return false
	}
	if l.Id != nil && r.Id != nil && *l.Id != *r.Id {
		return false
	}
	if l.ChildOf != nil && r.Id != nil && *l.ChildOf != *r.ChildOf {
		return false
	}
	if !bytes.Equal(l.PasswordHash, r.PasswordHash) {
		return false
	}
	return true
}

func cmpObject(obj []byte, link string) (bool, error) {
	resp, err := http.Get(link)
	if err != nil {
		return false, err
	}
	var buf bytes.Buffer
	_, err = io.Copy(&buf, resp.Body)
	if err != nil {
		return false, err
	}
	equal := bytes.Equal(buf.Bytes(), obj)
	return equal, nil
}

func TestUserRepo_StoreFetchUpdate(t *testing.T) {
	sampleAccount, err := account.NewUserAccountWithPassword("invitation@email.com", "myname", "mydisplay", "password", "personal@email.com")
	if err != nil {
		t.Error(err)
	}

	id, err := userRepo.Store(context.Background(), sampleAccount)
	if err != nil {
		t.Error(err)
	}
	// update id
	sampleAccount.Id = &id

	firstFetch, err := userRepo.FetchById(context.Background(), id)
	if err != nil {
		t.Error(err)
	}

	if !cmpUser(*sampleAccount, *firstFetch) {
		t.Errorf("%v is not equal to %v\n", *sampleAccount, *firstFetch)
	}

	secondFetch, err := userRepo.FetchByUsername(context.Background(), sampleAccount.Username)
	if err != nil {
		t.Error(err)
	}

	if !cmpUser(*sampleAccount, *secondFetch) {
		t.Errorf("%v is not equal to %v\n", *sampleAccount, *secondFetch)
	}

	err = sampleAccount.VerifyEmail(sampleAccount.UnverifiedEmail)
	if err != nil {
		t.Error(err)
	}
	err = userRepo.Update(context.Background(), sampleAccount)
	if err != nil {
		t.Error(err)
	}

	thirdFetch, err := userRepo.FetchByEmail(context.Background(), sampleAccount.Email)
	if err != nil {
		t.Error(err)
	}

	if len(thirdFetch) < 1 {
		t.Errorf("cannot find any user with email address: %v", sampleAccount.Email)
	}

	if !cmpUser(*sampleAccount, thirdFetch[0]) {
		t.Errorf("%v is not equal to %v\n", *sampleAccount, thirdFetch[0])
	}
}

func TestUserRepo_Exist(t *testing.T) {
	sampleAccount, err := account.NewUserAccountWithPassword("thefabulous@email.com", "thyname", "thydisplay", "password", "particle@email.com")
	if err != nil {
		t.Error(err)
	}

	id, err := userRepo.Store(context.Background(), sampleAccount)
	if err != nil {
		t.Error(err)
	}
	// update id
	sampleAccount.Id = &id

	exist, err := userRepo.HasUsername(context.Background(), sampleAccount.Username)
	if err != nil {
		t.Error(err)
	}

	if !exist {
		t.Errorf("failed to find user with username %v", sampleAccount.Username)
	}

	exist, err = userRepo.HasInvitationEmail(context.Background(), sampleAccount.InvitationEmail)
	if err != nil {
		t.Error(err)
	}

	if !exist {
		t.Errorf("failed to find user with invitation email %v", sampleAccount.InvitationEmail)
	}
}

func randomImage() []byte {
	width := 10
	height := 10

	upLeft := image.Point{0, 0}
	lowRight := image.Point{width, height}

	img := image.NewRGBA(image.Rectangle{upLeft, lowRight})
	
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			img.Set(x, y, color.RGBA{
				uint8(rand.Int() % 255),
				uint8(rand.Int() % 255),
				uint8(rand.Int() % 255),
				0xff,
			})
		}
	}

	var buf bytes.Buffer
	png.Encode(&buf, img)
	return buf.Bytes()
}

func TestBusinessRepo_StoreFetchUpdate(t *testing.T) {
	server := httptest.NewServer(objectMock.Handler())
	defer server.Close()
	objectMock.Address = server.URL

	sampleDoc := [][]byte{randomImage(), randomImage(), randomImage()}
	
	sampleIdentity, err := account.NewBusinessIdentity("officialName", "ANREG1000", "10, al, Prifthenas")
	if err != nil {
		t.Error(err)
	}
	
	sample, err := account.NewBusinessAccountWithPassword("mybusiness", "publicname", "password", "business@email.com", sampleIdentity)
	if err != nil {
		t.Error(err)
	}

	id, err := businessRepo.Store(context.Background(), sample, sampleDoc)
	if err != nil {
		t.Error(err)
	}

	sample.Id = &id

	firstFetch, err := businessRepo.FetchById(context.Background(), id)
	if err != nil {
		t.Error(err)
	}

	if !cmpBusiness(*sample, *firstFetch) {
		t.Errorf("%v is not equal to %v\n", *sample, *firstFetch)
	}

	for i, v := range sampleDoc {
		link := firstFetch.Documents[i]
		equal, err := cmpObject(v, link)
		if err != nil {
			t.Error(err)
		}
		if !equal {
			t.Errorf("url %v is not equal to bytes %v\n", link, v)
		}
	}

	sample = firstFetch

	secondFetch, err := businessRepo.FetchByUsername(context.Background(), sample.Username)
	if err != nil {
		t.Error(err)
	}

	if !cmpBusiness(*sample, *secondFetch) {
		t.Errorf("%v is not equal to %v\n", *sample, *secondFetch)
	}

	newDoc := [][]byte{randomImage()}

	err = sample.VerifyEmail(sample.UnverifiedEmail)
	if err != nil {
		t.Error(err)
	}
	err = businessRepo.Update(context.Background(), sample, newDoc)
	if err != nil {
		t.Error(err)
	}
	thirdFetch, err := businessRepo.FetchByEmail(context.Background(), sample.Email)
	if err != nil {
		t.Error(err)
	}

	if len(thirdFetch) < 1 {
		t.Errorf("cannot find any user with email address: %v", sample.Email)
	}

	if !cmpBusiness(*sample, thirdFetch[0]) {
		t.Errorf("%v is not equal to %v\n", *sample, thirdFetch[0])
	}

	for i, v := range append(sampleDoc, newDoc...) {
		link := thirdFetch[0].Documents[i]
		equal, err := cmpObject(v, link)
		if err != nil {
			t.Error(err)
		}
		if !equal {
			t.Errorf("url %v is not equal to bytes %v\n", link, v)
		}
	}
}

func TestBusinessRepo_Exist(t *testing.T) {
	server := httptest.NewServer(objectMock.Handler())
	defer server.Close()
	objectMock.Address = server.URL

	sampleDoc := [][]byte{randomImage(), randomImage(), randomImage()}
	
	sampleIdentity, err := account.NewBusinessIdentity("officialName", "ANREG1000", "10, al, Prifthenas")
	if err != nil {
		t.Error(err)
	}
	
	sample, err := account.NewBusinessAccountWithPassword("mybusiness", "publicname", "password", "business@email.com", sampleIdentity)
	if err != nil {
		t.Error(err)
	}

	id, err := businessRepo.Store(context.Background(), sample, sampleDoc)
	if err != nil {
		t.Error(err)
	}

	sample.Id = &id

	exist, err := businessRepo.HasUsername(context.Background(), sample.Username)
	if err != nil {	
		t.Error(err)
	}

	if !exist {
		t.Errorf("user %v should exist", sample.Username)
	}
}
