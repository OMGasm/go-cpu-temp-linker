package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	Cpu_sensor      string
	Cpu_input_label string
}

func read_cfg() (*Config, error) {
	file, err := os.Open("config.json")
	defer file.Close()
	if err != nil {
		return nil, err
	}

	dec := json.NewDecoder(file)
	var cfg Config
	err = dec.Decode(&cfg)
	return &cfg, err
}

func perr(msg string, err error) {
	fmt.Fprintln(os.Stderr, msg, err)
}

func main() {
	cfg, err := read_cfg()
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			cfg = &Config{
				Cpu_sensor:      "k10temp",
				Cpu_input_label: "Tctl",
			}
		} else {
			perr("Unable to read config for some reason\n", err)
		}
	}
	hw_dir := os.DirFS("/sys/class/hwmon")
	dirs, err := fs.ReadDir(hw_dir, ".")
	if err != nil {
		perr("can't read hwmon?", err)
		return
	}
	for _, dirmon := range dirs {
		info, _ := dirmon.Info()
		if info.Mode()&fs.ModeSymlink == 0 {
			continue
		}
		dev_fs, err := fs.Sub(hw_dir, dirmon.Name())
		if err != nil {
			perr("", err)
		}
		println(dirmon.Name())
		name_file, err := dev_fs.Open("name")
		if err != nil {
			perr("can't read device name?", err)
		}
		var name_bytes [64]byte
		size, err := name_file.Read(name_bytes[:])
		if err != nil {
			perr("can't read device name?", err)
		}
		name_file.Close()
		name := string(name_bytes[:size])
		name = strings.TrimSpace(name)

		if name == cfg.Cpu_sensor {
			labels, err := fs.Glob(dev_fs, "temp*_label")
			if err != nil {
				perr("i don't evne know", err)
			}
			for _, label := range labels {
				fmt.Println(name, label)
				label_file, err := dev_fs.Open(label)
				if err != nil {
					perr("label file bad?", err)
					continue
				}
				var label_bytes [32]byte
				size, err = label_file.Read(label_bytes[:])
				if err != nil {
					perr("err reading temp?", err)
					continue
				}
				label_file.Close()
				label_s := string(label_bytes[:size])
				label_s = strings.TrimSpace(label_s)
				fmt.Println(label_s)

				tinput := strings.TrimSuffix(label, "_label")

				path := filepath.Join("/sys/class/hwmon", dirmon.Name(), tinput+"_input")
				fmt.Println(tinput)
				fmt.Println(path)

				err = os.Remove("temp_input")
				if err != nil && !errors.Is(err, os.ErrNotExist) {
					perr("remove temp symlink", err)
				}
				err = os.Symlink(path, "temp_input")
				if err != nil {
					perr("symlink", err)
					continue
				}
				tempf, err := os.Open("temp_input")
				if err != nil {
					perr("temp open", err)
					continue
				}
				var temp_bytes [32]byte
				size, err := tempf.Read(temp_bytes[:])
				if err != nil {
					perr("temp read", err)
					break
				}
				temp := string(temp_bytes[:size])
				temp = strings.TrimSpace(temp)
				fmt.Println(temp)
			}
		}
		fmt.Println()
	}
}
