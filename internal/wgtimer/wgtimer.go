package wgtimer

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

// Queue requests a Windguru collector timer for stationID.
// A root cron job (deploy/process-wg-timer-queue.sh) picks up files from queueDir.
// Empty queueDir skips (local dev).
func Queue(queueDir string, stationID int) error {
	if queueDir == "" {
		return nil
	}
	if stationID <= 0 {
		return fmt.Errorf("invalid station id %d", stationID)
	}
	if err := os.MkdirAll(queueDir, 0o755); err != nil {
		return fmt.Errorf("create timer queue dir: %w", err)
	}
	path := filepath.Join(queueDir, strconv.Itoa(stationID))
	if err := os.WriteFile(path, []byte(strconv.Itoa(stationID)+"\n"), 0o644); err != nil {
		return fmt.Errorf("queue wg timer %d: %w", stationID, err)
	}
	return nil
}
