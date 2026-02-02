package config

import "strconv"

type ES struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
}

func (es *ES) URL() string {
	return es.Host + ":" + strconv.Itoa(es.Port)
}
