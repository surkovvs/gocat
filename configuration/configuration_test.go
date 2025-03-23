package configuration

import (
	"testing"
)

func TestParseFile(t *testing.T) {
	cfg, err := ParseFile(`test_data/config_1.yml`)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%+v", *cfg)
	t.Logf("%+v", *cfg.Logger.GetLogLvl())

	cfg, err = ParseFile(`test_data/config_2.yml`)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%+v", *cfg)
	t.Logf("%+v", *cfg.Logger.GetLogLvl())
	cfg, err = ParseFile(`test_data/config_3.yml`)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%+v", *cfg)
	t.Logf("%+v", cfg.Logger.GetLogLvl())
}
