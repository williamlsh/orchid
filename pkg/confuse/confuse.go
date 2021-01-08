package confuse

import (
	"errors"
	"math"

	"github.com/pjebs/optimus-go"
)

// don't change it
const (
	prime      = 961748951
	modInverse = 1870611431
	random     = 1045454339
)

var (
	op = optimus.Optimus{}
)

func init() {
	op = optimus.New(prime, modInverse, random)
}

// EncodeID
// id must be an integer less than math.MaxInt32 and greater than 0.
func EncodeID(id uint64) (midID uint64, err error) {
	if id < 0 || id > math.MaxInt32 {
		err = errors.New("id is lagger than max int32")
		return
	}
	midID = op.Encode(id)
	return
}

// DecodeID is used to decode n back to the original
// mixID must be an integer less than math.MaxInt32 and greater than 0.
func DecodeID(mixID uint64) (id uint64, err error) {
	if id < 0 || id > math.MaxInt32 {
		err = errors.New("mixID is larger than max int32")
		return
	}
	id = op.Decode(mixID)
	return
}
