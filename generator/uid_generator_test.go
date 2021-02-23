package generator

import (
	"fmt"
	"testing"
)

func TestDefaultUidGenerator_GetUID(t *testing.T) {
	fmt.Println("========== start generate ===========")
	generator := NewDefaultUidGenerator()
	uid, err := generator.GetUID()
	if err != nil {
		println(err)
	} else {
		println(uid)
	}
	fmt.Println("=========== end generate ==============")
}
