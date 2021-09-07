package sqltype

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"math/big"
)

type MyBool bool

func (b MyBool) Value() (driver.Value, error) {
	if b {
        return []byte{1}, nil
    } else {
        return []byte{0}, nil
    }
} 

func (b *MyBool) Scan(src interface{}) error {
	v, ok := src.([]byte)
	if !ok {
		return fmt.Errorf("Unexpected type for MyBool: %T", src)
	}
	*b = v[0] == 1
	return nil
}

func ToBigInt(s sql.NullString) *big.Int {
	n, _ := new(big.Int).SetString(s.String, 10)
	return n
}