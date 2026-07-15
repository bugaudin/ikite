package wgtimer

import (
	"fmt"
	"os/exec"
	"strconv"
)

// Enable starts (or ensures) the systemd timer for a Windguru station ID.
// script is the path to add-wg-timer.sh; empty skips (local dev).
func Enable(script string, stationID int) error {
	if script == "" {
		return nil
	}
	if stationID <= 0 {
		return fmt.Errorf("invalid station id %d", stationID)
	}
	cmd := exec.Command("sudo", script, strconv.Itoa(stationID))
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("enable wg timer %d: %w: %s", stationID, err, string(out))
	}
	return nil
}
