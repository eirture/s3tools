package config

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v2"
)

const (
	KB = 1024
	MB = 1024 * KB
	GB = 1024 * MB
)

type Config struct {
	Credential *Credential `yaml:"credential"`

	Bucket          string   `yaml:"bucket"`
	FileSizeList    []string `yaml:"file_size_list"`
	Workers         int      `yaml:"workers"`
	DeleteAfterDays []string `yaml:"delete_after_days"`
	PartSize        int64    `yaml:"part_size"`

	FileSizes []int64 `yaml:"-"`
}

func (c *Config) verify() (err error) {
	var errs []string
	if len(c.FileSizeList) == 0 {
		errs = append(errs, "no file_size_list")
	}
	if len(c.DeleteAfterDays) == 0 {
		errs = append(errs, "no delete after days")
	}

	if c.Workers <= 0 {
		c.Workers = 1
	}

	if len(errs) == 0 {
		return nil
	}
	return fmt.Errorf("invalid config. %s", strings.Join(errs, ", "))
}

func Load(path string) (cfg *Config, err error) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		return
	}
	cfg = &Config{}
	if err = yaml.Unmarshal(data, cfg); err != nil {
		return
	}

	if err = cfg.verify(); err != nil {
		return
	}

	sizeList := make([]int64, len(cfg.FileSizeList))
	var unknownSizes []string
	for i, fs := range cfg.FileSizeList {
		n := len(fs)
		fstr := fs[:n-1]
		basic := int64(64)
		switch fs[n-1] {
		case 'G':
			basic = GB
		case 'M':
			basic = MB
		case 'K':
			basic = KB
		default:
			fstr = fs
		}
		basicSize, err := strconv.ParseInt(fstr, 10, 64)
		if err != nil {
			unknownSizes = append(unknownSizes, fs)
		}
		sizeList[i] = basicSize * basic
	}
	cfg.FileSizes = sizeList

	return
}
