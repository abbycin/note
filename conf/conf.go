/***********************************************
        File Name: conf
        Author: Abby Cin
        Mail: abbytsing@gmail.com
        Created Time: 10/26/19 10:30 AM
***********************************************/

package conf

import (
	"blog/logging"
	"github.com/BurntSushi/toml"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"
)

type Config struct {
	Blacklist Blacklist `toml:"blacklist"`
	Common    Common    `toml:"common"`
	Session   Session   `toml:"session"`
	Model     Model     `toml:"model"`
	Service   Service   `tmol:"service"`
}

type Blacklist struct {
	Expiry int64 `toml:"expiry"`
	Limit  int   `toml:"limit"`
}

type Common struct {
	Addr      string  `toml:"addr"`
	Assets    string  `toml:"assets"`
	Images    string  `toml:"images"`
	DbFile    string  `toml:"db_file"`
	Logging   Logging `toml:"logging"`
	ProxyMode bool    `toml:"proxy_mode"`
}

type Logging struct {
	Stderr       bool          `toml:"stderr"`
	FileName     string        `toml:"filename"`
	RollSize     int64         `toml:"roll_size"`
	RollInterval time.Duration `toml:"roll_interval"`
	Level        string        `toml:"level"`
}

type Session struct {
	Expiry  int    `toml:"expiry"`
	Key     string `toml:"key"`
	AuthKey string `toml:"auth_key"`
	AuthVal string `toml:"auth_val"`
}

type Model struct {
	Title    string          `toml:"title"`
	TmplRoot string          `toml:"tmpl_root"`
	Edit     ModelData       `toml:"edit"`
	Login    ModelData       `toml:"login"`
	Manage   ManageModelData `toml:"manage"`
	Home     ModelData       `toml:"home"`
	Article  ModelData       `toml:"article"`
}

type ManageModelData struct {
	MainApi string `toml:"main_api"`
	SubApi  string `toml:"sub_api"`
	Main    string `toml:"main"`
	Setting string `toml:"setting"`
}

type ModelData struct {
	Api  string `toml:"api"`
	Tmpl string `toml:"tmpl"`
}

type Service struct {
	Assets string `toml:"assets"`
	Login  string `toml:"login"`
	Edit   string `toml:"edit"`
	Image  string `toml:"image"`
	Manage string `toml:"manage"`
	Navi   string `toml:"navi"`
	User   string `toml:"user"`
}

func (c *Config) str2Level() int {
	l := strings.ToLower(c.Common.Logging.Level)
	switch l {
	case "info":
		return logging.INFO
	case "debug":
		return logging.DEBUG
	case "warn":
		return logging.WARN
	case "error":
		return logging.ERROR
	case "fatal":
		return logging.FATAL
	default:
		panic("invalid level " + c.Common.Logging.Level)
	}
}

func (c *Config) GetLogger() logging.Config {
	return logging.Config{
		RollSize:     c.Common.Logging.RollSize,
		RollInterval: c.Common.Logging.RollInterval,
		FileName:     c.Common.Logging.FileName,
		Level:        c.str2Level(),
	}
}

func Init(cfgPath string) *Config {
	f, err := os.OpenFile(cfgPath, os.O_RDONLY, 00644)
	if err != nil {
		log.Panicf("can't open configuration file: %s\n", err)
	}
	defer f.Close()
	data, err := ioutil.ReadAll(f)
	if err != nil {
		log.Panicf("can't read configuration file: %s\n", err)
	}
	res := new(Config)
	_, err = toml.Decode(string(data), res)
	if err != nil {
		log.Panicf("can't decode configuration file: %s\n", err)
	}
	return res
}
