package hashidsx

import (
	"errors"

	"github.com/speps/go-hashids"
)

const (
	salt      = "this my salt."
	minLength = 30
)

var ErrInvalidHashID = errors.New("invalid hash string, more than one id is encoded")

var hd *hashids.HashID

func init() {
	hd = hashid()
}

// Encode encodes a single integer id to string.
func Encode(id int) (string, error) {
	return hd.Encode([]int{id})
}

// Decode decodes a hash string to a single integer id.
// If the decoded hash is not a single id, it returns an error indicating this.
func Decode(hash string) (int, error) {
	d, err := hd.DecodeWithError(hash)
	if err != nil {
		return 0, err
	}
	if len(d) > 1 {
		return 0, ErrInvalidHashID
	}
	return d[0], nil
}

func hashid() *hashids.HashID {
	hd := hashids.NewData()
	hd.Salt = salt
	hd.MinLength = minLength

	return hashids.NewWithData(hd)
}
