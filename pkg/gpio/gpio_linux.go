// +build linux

package gpio

import (
	"context"
	"log"
	"sync"
	"time"

	// This is the older GPIO library which is required for Linux kernel
	// versions older than v5.5 https://github.com/warthog618/gpio
	wgpio "github.com/warthog618/gpio"

	"github.com/wwwil/launchpoint/pkg/launchpoint"
)

// TODO: Use GPIO new style:
//       https://github.com/warthog618/gpiod#edge-watches
// TODO: Check kernel version and use correct GPIO library:
//       https://github.com/zcalusic/sysinfo
//       https://stackoverflow.com/questions/53197586/how-can-i-detect-os-version-in-go
// TODO: Implement GPIO output based on poll HTTP
//       https://stackoverflow.com/questions/39363794/periodically-polling-a-rest-endpoint-in-go

// watchPollPeriodMilliseconds is how long the watch function will wait for
// between pin readings.
const watchPollPeriodMilliseconds = 100

// Run sets up GPIO and starts watching for triggers.

func Run(ctx context.Context, wg *sync.WaitGroup, config *launchpoint.Config) {
	// Defer calls are added to a stack, executed in last in first out order, so
	// this will be called last, after the deferred calls to unwatch the GPIO
	// pins.
	defer log.Println("GPIO stopped.")
	defer wg.Done()

	if len(config.GPIOTriggers) < 1 {
		log.Println("no GPIO triggers set")
		return
	}
	log.Println("starting GPIO triggers")

	// Open memory range for GPIO access in /dev/gpiomem.
	err := wgpio.Open()
	if err != nil {
		log.Println(err)
		return
	}
	defer wgpio.Close()

	// Watch for GPIO triggers.
	for i := range config.GPIOTriggers {
		go watch(ctx, &config.GPIOTriggers[i])
	}

	// Run until told to stop.
	<-ctx.Done()
}

// watch polls a GPIO pin at a set interval and when triggered makes any
// requests associated with it.
func watch(ctx context.Context, trigger *launchpoint.GPIOTrigger) {
	// The GPIO library used has it's own watch functionality, however during
	// testing this seemed to have issues with repeatedly being triggered for a
	// single input from a button press and doesn't offer any debounce options.
	log.Printf("watching pin %d", trigger.Pin)

	pin := wgpio.NewPin(trigger.Pin)
	pin.Input()
	pin.PullUp()
	defer pin.PullNone()
	triggered := false
	for {
		// The pins are pulled up so watch for them going low.
		if pin.Read() == wgpio.Low {
			// If the pin was already triggered then do nothing.
			if triggered {
				continue
			}
			// Set the pin to triggered and print a message.
			triggered = true
			log.Printf("pin %d pressed", pin.Pin())
			// Make any requests specified for this trigger.
			for _, request := range trigger.Requests {
				err := request.Make()
				if err != nil {
					log.Println(err)
				}
			}
		} else {
			// Reset the pin.
			triggered = false
		}
		time.Sleep(watchPollPeriodMilliseconds * time.Millisecond)
	}
}


