package hashidsx

import (
	"errors"
	"fmt"
	"testing"
)

func TestHashID(t *testing.T) {
	t.Run("one id", func(t *testing.T) {
		id := 7
		hash, err := Encode(id)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println("encoded: ", hash)

		decoded, err := Decode(hash)
		if err != nil {
			t.Fatal(err)
		}
		if decoded != id {
			t.Fatalf("expect %d, got %d", id, decoded)
		}
	})
	t.Run("more than one id", func(t *testing.T) {
		e, err := hd.Encode([]int{45, 434, 1313, 99})
		if err != nil {
			t.Fatal(err)
		}

		_, err = Decode(e)
		if !errors.Is(ErrInvalidHashID, err) {
			t.Fatalf("not this error: %v", err)
		}
	})
}
