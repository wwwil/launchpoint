// +build darwin

// TODO: Consider removing this and just building for linux.

package gpio

import (
	"context"
	"log"
	"sync"

	"github.com/wwwil/launchpoint/pkg/launchpoint"
)

func Run(ctx context.Context, wg *sync.WaitGroup, cfg *launchpoint.Config) {
	defer wg.Done()
	if len(cfg.GPIOTriggers) > 0 {
		log.Println("Launchpoint does not support GPIO on this platform.")
	}
	return
}
