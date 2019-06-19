package config

import (
	"github.com/BurntSushi/toml"
	"path/filepath"
)

type Ips map[string][]string

func Read(path string) (c Ips,e error) {
	cfg := new(Ips)
	p,_ := filepath.Abs(path)
	if _, err := toml.DecodeFile(p,&cfg); err != nil{
		return nil,err
	}
	return *cfg,nil
}