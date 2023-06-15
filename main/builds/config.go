package builds

import (
	"XrayHelper/main/errors"
	"XrayHelper/main/log"
	"bufio"
	"github.com/creasty/defaults"
	"gopkg.in/yaml.v3"
	"os"
	"strings"
)

const packageListPath = "/data/system/packages.list"

var ConfigFilePath *string
var PackageMap = make(map[string]string)

// Config the program configuration, yml
var Config struct {
	XrayHelper struct {
		CoreType   string   `default:"xray" yaml:"coreType"`
		CorePath   string   `yaml:"corePath"`
		CoreConfig string   `yaml:"coreConfig"`
		DataDir    string   `yaml:"dataDir"`
		RunDir     string   `yaml:"runDir"`
		ProxyTag   string   `default:"proxy" yaml:"proxyTag"`
		SubList    []string `yaml:"subList"`
	} `yaml:"xrayHelper"`
	Proxy struct {
		Method       string   `default:"tproxy" yaml:"method"`
		TproxyPort   string   `default:"65535" yaml:"tproxyPort"`
		SocksPort    string   `default:"65534" yaml:"socksPort"`
		ClashDNSPort string   `default:"65533" yaml:"clashDNSPort"`
		EnableIPv6   bool     `default:"false" yaml:"enableIPv6"`
		Mode         string   `default:"blacklist" yaml:"mode"`
		PkgList      []string `yaml:"pkgList"`
		ApList       []string `yaml:"apList"`
		IgnoreList   []string `yaml:"ignoreList"`
		IntraList    []string `yaml:"intraList"`
	} `yaml:"proxy"`
	Clash struct {
		OverrideConfig string `yaml:"overrideConfig"`
		DNSPort        string `default:"65533" yaml:"dnsPort"`
		ApiListen      string `yaml:"apiListen"`
		ExternalUIDir  string `yaml:"externalUIDir"`
	} `yaml:"clash"`
}

// LoadConfig load program configuration file, should be called before any command Execute
func LoadConfig() error {
	configFile, err := os.ReadFile(*ConfigFilePath)
	if err != nil {
		return errors.New("load config failed, ", err).WithPrefix("config")
	}
	if err := defaults.Set(&Config); err != nil {
		return errors.New("set default config failed, ", err).WithPrefix("config")
	}
	if err := yaml.Unmarshal(configFile, &Config); err != nil {
		return errors.New("unmarshal config failed, ", err).WithPrefix("config")
	}
	log.HandleDebug(Config.XrayHelper)
	log.HandleDebug(Config.Proxy)
	return nil
}

// LoadPackage load and parse Android package with uid list into a map
func LoadPackage() error {
	packageListFile, err := os.Open(packageListPath)
	if err != nil {
		return errors.New("load package failed, ", err).WithPrefix("config")
	}
	packageScanner := bufio.NewScanner(packageListFile)
	packageScanner.Split(bufio.ScanLines)
	for packageScanner.Scan() {
		packageInfo := strings.Fields(packageScanner.Text())
		if len(packageInfo) >= 2 {
			PackageMap[packageInfo[0]] = packageInfo[1]
		}
	}
	if err := packageListFile.Close(); err != nil {
		return errors.New("close package file failed, ", err).WithPrefix("config")
	}
	log.HandleDebug(PackageMap)
	return nil
}
