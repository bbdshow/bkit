package ordernum

import (
	"fmt"
	"testing"
)

func TestNewOrderId(t *testing.T) {
	orderId := NewOrderId().WithTag("m").WithTag("m")
	fmt.Println(orderId.String(), orderId.Time(), orderId.Tags(), orderId.ExistsTag("m"))
}
