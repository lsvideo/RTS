// config_test
package main

import (
	"testing"
)

func TestParse_config(t *testing.T) {
	var config SL_config
	err := Parse_config("./sl.yaml", &config)
	if err != nil {
		t.Errorf("Parse config failed! %v", err)
	} else {
		t.Logf("%v", config)
	}

}
