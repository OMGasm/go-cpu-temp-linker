package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

func perr(msg string, err error) {
	fmt.Fprintln(os.Stderr, msg, err)
}

func main() {
	cfg, err := Read_config()
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			cfg = Default_Config()
		} else {
			perr("Unable to read config for some reason\n", err)
		}
	}

	hw_dir := "/sys/class/hwmon"
	devices, err := Enumerate_sensors(hw_dir)
	if err != nil {
		perr("can't enumerate hwmon?", err)
		return
	}

	var device *Sensor
	for dev := range devices {
		if dev.Name == cfg.Cpu_sensor {
			device = &dev
			break
		}
	}
	if device == nil {
		fmt.Printf("Unable to find hwmon device for %s\n", cfg.Cpu_sensor)
		return
	}

	probes, err := Enumerate_probes(device)
	if err != nil {
		perr("Error enumerating probes", err)
	}
	var probe *Probe
	for p := range probes {
		if p.Label == cfg.Cpu_input_label {
			probe = &p
			break
		}
	}
	if probe == nil {
		fmt.Printf("Unable to find probe for %s in %s", cfg.Cpu_input_label, cfg.Cpu_sensor)
		return
	}

	err = Create_hwmon_symlink(filepath.Join(device.Path, probe.Input), "temp_input")
	if err != nil {
		perr("Error creating symlink", err)
		return
	}
	_, err = Read_hwmon_file("temp_input")
	if err != nil {
		perr("Error reading temp from symlink", err)
		return
	}
}
