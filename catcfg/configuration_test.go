package catcfg

import (
	"testing"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

func TestParseFile(t *testing.T) {
	cfg, err := ParseFile(`test_data/config_1.yml`)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%+v", *cfg)
	t.Logf("%+v", (*cfg).Kafka)
	// t.Logf("%+v", (*cfg).Kafka["produser"])
	t.Logf("%+v", *cfg.GetLogLvl())
	prodCfg := (*cfg).Kafka["produser"]
	if _, err := kafka.NewConsumer(&prodCfg); err != nil {
		t.Fatal(err)
	}

	// cfg, err = ParseFile(`test_data/config_2.yml`)
	// if err != nil {
	// 	t.Fatal(err)
	// }
	// t.Logf("%+v", *cfg)
	// t.Logf("%+v", *cfg.GetLogLvl())
	// cfg, err = ParseFile(`test_data/config_3.yml`)
	// if err != nil {
	// 	t.Fatal(err)
	// }
	// t.Logf("%+v", *cfg)
	// t.Logf("%+v", cfg.GetLogLvl())
}
