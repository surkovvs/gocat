package appsim

import (
	"errors"
	"time"

	"github.com/surkovvs/ggwp"
)

func Init() {
	ggwp.GetDefault().RegisterNamedGracefulStop("first", func() error {
		time.Sleep(time.Second / 2)
		return nil
	})
	ggwp.GetDefault().RegisterNamedGracefulStop("second", func() error {
		time.Sleep(time.Second * 2)
		return nil
	})
	ggwp.GetDefault().RegisterNamedGracefulStop("third", func() error {
		return errors.New("error from third")
	})
}
