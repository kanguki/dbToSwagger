package openApi

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
)

var db Db

func TestMain(m *testing.M) {
	conn, err := sql.Open("mysql", "root:password@tcp(localhost)/tradex-configuration")
	if err != nil {
		Log("error connecting db: %v", err)
		return
	}
	Log("Successfully connected mysql")
	defer conn.Close()
	db = &OpenApiDb{Conn: conn}
	m.Run()
}

func TestRead(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping TestRead in short mode")
	}
	var data = make(chan RawData, 100)
	go func() {
		for v := range data {
			t.Log(v)
		}
	}()
	err := db.Read(data, DBOptions{ClientId: "kis-wts", Domain: "kis"})
	assert.NoError(t, err)

}

func TestHasSuspiciousCharacter(t *testing.T) {
	shouldTrue := []string{"\"dangerous", "' and \\'"}
	for _, v := range shouldTrue {
		assert.Equal(t, hasSuspiciousCharacter(v), true, v)
	}
	shouldFalse := []string{"_abc", " xyz", "a"}
	for _, v := range shouldFalse {
		assert.Equal(t, hasSuspiciousCharacter(v), false, v)
	}
}
