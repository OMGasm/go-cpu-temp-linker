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

	hw_dir := "/sys/class/hwmon"
	dirs, err := os.ReadDir(hw_dir)
	if err != nil {
		perr("can't read hwmon?", err)
		return
	}
	for _, dirmon := range dirs {
		info, _ := dirmon.Info()
		if info.Mode()&fs.ModeSymlink == 0 {
			continue
		}
		dev_path := filepath.Join(hw_dir, dirmon.Name())
		name_file, err := os.Open(filepath.Join(dev_path, "name"))
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
			dev_fs := os.DirFS(dev_path)
			labels, err := fs.Glob(dev_fs, "temp*_label")
			if err != nil {
				perr("i don't evne know", err)
			}
			for _, label_path := range labels {
				label, err := read_hwmon_file(filepath.Join(dev_path, label_path))
				if err != nil {
					perr("Failed to read probe label", err)
					continue
				}

				if label != cfg.Cpu_input_label {
					continue
				}

				tinput := strings.TrimSuffix(label_path, "_label")

				path := filepath.Join(dev_path, tinput+"_input")

				fmt.Printf("Found probe [%s/%s] at %s\n", name, label, path)

				err = create_hwmon_symlink(path, "temp_input")
				if err != nil {
					perr("Failed to create probe symlink", err)
					return
				}

				_, err = read_hwmon_file("temp_input")
				if err != nil {
					perr("Failed to read probe symlink", err)
					return
				}
				fmt.Printf("Symlinked probe to %s\n", "temp_input")
			}
		}
	}
}

func create_hwmon_symlink(target_path string, link_path string) error {
	err := os.Remove(link_path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		perr("Failed to remove probe symlink", err)
		return err
	}
	err = os.Symlink(target_path, link_path)
	return err
}

func read_hwmon_file(path string) (string, error) {
	file, err := os.Open(path)
	defer file.Close()
	if err != nil {
		return "", nil
	}

	var file_bytes [32]byte
	size, err := file.Read(file_bytes[:])
	if err != nil {
		return "", err
	}

	str := string(file_bytes[:size])
	str = strings.TrimSpace(str)

	return str, nil
}
