package openApi

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func BenchmarkRead(b *testing.B) {
	if testing.Short() {
		b.Skip("skipping TestRead in short mode")
	}
	for n := 0; n < b.N; n++ {
		var data = make(chan RawData, 100)
		db.Read(data, DBOptions{ClientId: "kis-wts", Domain: "kis"})
	}
	// for v := range data {
	// 	t.Log(v)
	// }

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
