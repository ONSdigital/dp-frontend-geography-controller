package instance

import (
	"context"
	"fmt"
	// "strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

// Reader used to trigger errors on reading
type Reader struct {
}

// Read a function used to mock errors
func (f Reader) Read(bytes []byte) (int, error) {
	return 0, fmt.Errorf("Reader failed")
}

var ctx = context.Background()

func TestSomething(t *testing.T) {
	Convey("Given some integer with a starting value", t, func() {
		x := 1

		Convey("When the integer is incremented", func() {
			x++

			Convey("The value should be greater by one", func() {
				So(x, ShouldEqual, 2)
			})
		})
	})
}
