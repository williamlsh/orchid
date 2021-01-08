package confuse

import (
	"math"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/willf/bitset"
)

func TestEncodeID_DecodeID(t *testing.T) {
	r := rand.New(rand.NewSource(time.Now().Unix()))
	for i:=0;i<100000;i++ {
		randomID := r.Int31()
		mixID,err := EncodeID(randomID)
		assert.NoError(t, err)
		restoreID,err := DecodeID(mixID)
		assert.NoError(t, err)
		assert.Equal(t, randomID, restoreID)
	}
}

func TestDataUniqueness(t *testing.T) {
	t.Log("it could last two minutes here..............")
	b := bitset.New(math.MaxInt32)
	for i:=int32(0);i<math.MaxInt32;i++ {
		mixID,_ := EncodeID(i)
		if b.Test(uint(mixID)) {
			t.FailNow()
		} else {
			b.Set(uint(mixID))
		}
	}
	assert.Equal(t, uint32(math.MaxInt32), uint32(b.Count()))
}