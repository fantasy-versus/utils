package types_test

import (
	"fmt"
	"testing"

	"github.com/fantasy-versus/utils/types"
)

func TestNew(t *testing.T) {
	kk := types.SqlUuid{}
	kk.New()

	fmt.Printf("kk: %v\n", kk)
}
