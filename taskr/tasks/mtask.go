package tasks

import (
	"io"
	"time"

	"github.com/influx6/faux/utils"
)

// MasterTask provides higher level structure which provides a series of tasks
// which would be run in order where the main task is allowed a consistent hold on
// the input and output writers.
// Before and After tasks cant not down the calls, they are given a maximum of
// 5min and then killed.
type MasterTask struct {
	Main            *Task   `json:"main"`
	MaxRunTime      string  `json:"max_runtime"`
	MaxRunCheckTime string  `json:"max_checktime"`
	Before          []*Task `json:"before"`
	After           []*Task `json:"after"`
}

// Stop ends all it's internal tasks.
func (mt *MasterTask) Stop(m io.Writer) {

	// Stop the before tasks.
	for _, tk := range mt.Before {
		if tk.Stopped() {
			continue
		}

		tk.Stop(m)
	}

	mt.Main.Stop(m)

	// Stop the after tasks.
	for _, tk := range mt.After {
		if tk.Stopped() {
			continue
		}

		tk.Stop(m)
	}
}

// Run executes the givin master tasks in the other expected, passing the
// provided writer to collect all responses.
func (mt *MasterTask) Run(mout, merr io.Writer) error {
	runtimes, err := utils.GetDuration(mt.MaxRunTime)
	if err != nil {
		return err
	}

	checkTimes, err := utils.GetDuration(mt.MaxRunCheckTime)
	if err != nil {
		return err
	}

	// Execute the before tasks.
	for _, tk := range mt.Before {
		tk.EndCheck = checkTimes

		go func(tm *Task) {
			<-time.After(runtimes)

			if !tk.Stopped() {
				tk.Stop(mout)
			}
		}(tk)

		tk.Run(mout, merr)
	}

	// Set the check time.
	mt.Main.EndCheck = checkTimes

	// Execute the main tasks and allow it hold io.
	mt.Main.Run(mout, merr)

	// Execute the after tasks.
	for _, tk := range mt.After {
		tk.EndCheck = checkTimes

		go func(tm *Task) {
			<-time.After(runtimes)

			if !tk.Stopped() {
				tk.Stop(mout)
			}
		}(tk)

		tk.Run(mout, merr)
	}

	return nil
}
