package config

import "strconv"

type ES struct {
	Host     string `yaml:"host" json:"host"`
	Port     int    `yaml:"port" json:"port"`
	User     string `yaml:"user" json:"user"`
	Password string `yaml:"password" json:"password"`
}

func (es *ES) URL() string {
	return es.Host + ":" + strconv.Itoa(es.Port)
}
