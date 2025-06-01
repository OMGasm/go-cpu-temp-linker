package main

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type Sensor struct {
	Path string
	Name string
}

type Probe struct {
	Label string
	Input string
}

func Enumerate_sensors(hw_path string) (func(func(Sensor) bool), error) {
	hwmon_dirs, err := os.ReadDir(hw_path)
	if err != nil {
		return nil, err
	}

	iter := func(yield func(Sensor) bool) {
		for _, hw_dir := range hwmon_dirs {
			info, _ := hw_dir.Info()
			// afaik all hwmon dirs are symlinks
			// skip anything else; could be wrong
			if info.Mode()&fs.ModeSymlink == 0 {
				continue
			}

			hwmon_path := filepath.Join(hw_path, hw_dir.Name())
			name_path := filepath.Join(hwmon_path, "name")
			name, err := Read_hwmon_file(name_path)
			if err != nil {
				// I don't think this should be able to happen?
				perr("Could not read hwmon device name", err)
				return
			}

			dev := Sensor{
				Path: hwmon_path,
				Name: name,
			}

			if !yield(dev) {
				break
			}
		}
	}
	return iter, nil
}

func Enumerate_probes(sensor *Sensor) (func(func(Probe) bool), error) {
	dev_fs := os.DirFS(sensor.Path)
	labels, err := fs.Glob(dev_fs, "temp*_label")
	if err != nil {
		perr("i don't evne know", err)
	}

	iter := func(yield func(Probe) bool) {

		for _, label_path := range labels {
			label, err := Read_hwmon_file(filepath.Join(sensor.Path, label_path))
			if err != nil {
				//Probably also should not be able to happen
				perr("Could not read probe label", err)
			}
			input := strings.TrimSuffix(label_path, "_label")
			input = input + "_input"

			probe := Probe{
				Label: label,
				Input: input,
			}

			if !yield(probe) {
				break
			}
		}
	}
	return iter, nil
}

func Create_hwmon_symlink(target_path string, link_path string) error {
	err := os.Remove(link_path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		perr("Failed to remove probe symlink", err)
		return err
	}
	err = os.Symlink(target_path, link_path)
	return err
}

func Read_hwmon_file(path string) (string, error) {
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
