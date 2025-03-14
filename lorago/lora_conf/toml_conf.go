package lora_conf

import (
	"flag"
	"github.com/BurntSushi/toml"
	"github.com/LorraineWen/lorago/lora_log"
	"os"
)

var TomlConf = &Conf{
	Log:      make(map[string]any),
	Template: make(map[string]any),
	Db:       make(map[string]any),
	Pool:     make(map[string]any),
}

func init() {
	loadToml()
}
func loadToml() {
	confFile := flag.String("conf", "conf/cmd.toml", "app config file")
	flag.Parse()
	if _, err := os.Stat(*confFile); err != nil {
		lora_log.NewLogger().Info("conf/cmd.toml文件不存在")
		return
	}
	_, err := toml.DecodeFile(*confFile, TomlConf)
	if err != nil {
		lora_log.NewLogger().Info("conf/cmd.toml格式不对")
		return
	}
}
