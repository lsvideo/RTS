// sl_log_test
package sl_log

import (
	"testing"
)

func TestLog_init(t *testing.T) {
	log := Log_init()
	if log == nil {
		t.Errorf("SL log init failed!")

	}
}
