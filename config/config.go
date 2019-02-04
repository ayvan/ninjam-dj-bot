package config

import (
	"flag"
	"github.com/luci/go-render/render"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
)

type AppConfig struct {
	AppPath       string
	AppConfigPath string
	AppPidPath    string
	HTTPPort      string       `yaml:"http_port"`
	TracksDir     string       `yaml:"tracks_dir"`
	DBFile        string       `yaml:"db_file"`
	LV2HostConfig string       `yaml:"lv2host_config"`
	DaemonMode    bool         `yaml:"daemon"`
	AppName       string       `yaml:"app_name"`
	LogFile       string       `yaml:"log_file"`
	LogLevel      string       `yaml:"log_level"`
	Server        NinJamServer `yaml:"server"`
	Player        Player       `yaml:"player"`
}

type NinJamServer struct {
	Host         string `yaml:"host"`
	Port         string `yaml:"port"`
	Anonymous    bool   `yaml:"anonymous"`
	UserName     string `yaml:"user_name"`
	UserPassword string `yaml:"user_password"`
}

type Player struct {
	Dir     string `yaml:"dir"`
	Command string `yaml:"command"`
	Args    string `yaml:"args"`
}

var appConfig *AppConfig

func Init() {
	appConfig = &AppConfig{}

	workPath, _ := os.Getwd()
	workPath, _ = filepath.Abs(workPath)
	// initialize default configurations
	appConfig.AppPath, _ = filepath.Abs(filepath.Dir(os.Args[0]))

	strPtr := flag.String("c", "config.yaml", "config path")
	strPtrPid := flag.String("p", "", "pid path")

	flag.Parse()

	appConfig.AppPidPath = *strPtrPid

	appConfig.AppConfigPath = *strPtr

	if workPath != appConfig.AppPath {
		if FileExists(appConfig.AppConfigPath) {
			os.Chdir(appConfig.AppPath)
		} else {
			appConfig.AppConfigPath = filepath.Join(workPath, "config.yaml")
		}
	}

	if appConfig.HTTPPort == "" {
		appConfig.HTTPPort = "8080"
	}
	appConfig.DaemonMode = false
	appConfig.AppName = "ninjam-dj-bot"
	appConfig.LogFile = "stdout"

	content, err := ioutil.ReadFile(appConfig.AppConfigPath)
	if err != nil {
		logrus.Fatalf("Can`t read config file (%s): %v\n", appConfig.AppConfigPath, err)
	}

	err = yaml.Unmarshal(content, appConfig)
	if err != nil {
		logrus.Fatalf("Yaml file %s parsing error: %v", appConfig.AppConfigPath, err)
	}

	setLogger(appConfig.LogLevel, appConfig.LogFile)
	if !appConfig.DaemonMode {
		logrus.Info("Config loaded:", render.Render(appConfig))
	}

	runtime.GOMAXPROCS(runtime.NumCPU())
}

func setLogger(level, dest string) {
	lvl, err := logrus.ParseLevel(level)

	if err != nil {
		logrus.Fatalf("Unable to parse '%v' as a log level", level)
	}

	logrus.SetLevel(lvl)

	if dest != "stdout" {
		absDest, err := filepath.Abs(dest)
		if err != nil {
			logrus.Fatalf("Unable to get absolute file path %s: err: %s", dest, err)
		}

		out, err := os.OpenFile(absDest, os.O_CREATE|os.O_WRONLY, 0777)
		if err != nil {
			logrus.Fatalf("Unable to open file %s: err: %s", dest, err)
		}

		logrus.SetOutput(out)
	}

	return
}

// FileExists reports whether the named file or directory exists.
func FileExists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func Get() *AppConfig {
	return appConfig
}
