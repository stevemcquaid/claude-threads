package utils

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
)

// BatteryStatus contains the current battery state.
type BatteryStatus struct {
	Percentage int
	Charging   bool
}

// GetBatteryStatus returns the current battery status, or nil if unavailable.
func GetBatteryStatus() (*BatteryStatus, error) {
	switch runtime.GOOS {
	case "darwin":
		return getMacBattery()
	case "linux":
		return getLinuxBattery()
	default:
		return nil, nil
	}
}

// FormatBatteryStatus returns a formatted battery string, or nil if unavailable.
func FormatBatteryStatus() (*string, error) {
	status, err := GetBatteryStatus()
	if err != nil || status == nil {
		return nil, err
	}
	var s string
	if status.Charging {
		s = "🔌 AC"
	} else {
		s = fmt.Sprintf("🔋 %d%%", status.Percentage)
	}
	return &s, nil
}

var macBatteryPctRe = regexp.MustCompile(`(\d+)%`)
var macChargingRe = regexp.MustCompile(`(?i)(charging|AC attached)`)

func getMacBattery() (*BatteryStatus, error) {
	out, err := exec.Command("pmset", "-g", "batt").Output()
	if err != nil {
		return nil, nil
	}
	s := string(out)
	m := macBatteryPctRe.FindStringSubmatch(s)
	if m == nil {
		return nil, nil
	}
	pct, _ := strconv.Atoi(m[1])
	charging := macChargingRe.MatchString(s)
	return &BatteryStatus{Percentage: pct, Charging: charging}, nil
}

func getLinuxBattery() (*BatteryStatus, error) {
	for _, name := range []string{"BAT0", "BAT1", "battery"} {
		base := "/sys/class/power_supply/" + name
		capBytes, err := os.ReadFile(base + "/capacity")
		if err != nil {
			continue
		}
		pct, err := strconv.Atoi(strings.TrimSpace(string(capBytes)))
		if err != nil {
			continue
		}
		charging := false
		if statusBytes, err := os.ReadFile(base + "/status"); err == nil {
			status := strings.TrimSpace(string(statusBytes))
			charging = status == "Charging" || status == "Full"
		}
		return &BatteryStatus{Percentage: pct, Charging: charging}, nil
	}
	return nil, nil
}
