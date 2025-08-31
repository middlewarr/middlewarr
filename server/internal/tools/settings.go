package tools

import (
	"sync"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

var onceSettings sync.Once

var conf = koanf.Conf{
	Delim: ".",
}
var k = koanf.NewWithConf(conf)

func GetSettings() *koanf.Koanf {
	l := GetLogger()

	onceSettings.Do(func() {
		yamlPath := "data/settings.yml"
		if err := k.Load(file.Provider(yamlPath), yaml.Parser()); err != nil {
			l.Fatal().
				Err(err).
				Msg("error loading config")
		}
	})

	return k
}
