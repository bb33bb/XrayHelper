package commands

import (
	"XrayHelper/main/builds"
	"XrayHelper/main/common"
	"XrayHelper/main/errors"
	"XrayHelper/main/log"
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"encoding/json"
	"io"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	singboxUrl           = "https://api.github.com/repos/SagerNet/sing-box/releases/latest"
	xrayCoreDownloadUrl  = "https://github.com/XTLS/Xray-core/releases/latest/download/Xray-android-arm64-v8a.zip"
	geoipDownloadUrl     = "https://github.com/Loyalsoldier/v2ray-rules-dat/releases/latest/download/geoip.dat"
	geositeDownloadUrl   = "https://github.com/Loyalsoldier/v2ray-rules-dat/releases/latest/download/geosite.dat"
	tun2socksDownloadUrl = "https://github.com/heiher/hev-socks5-tunnel/releases/latest/download/hev-socks5-tunnel-linux-arm64"
)

type UpdateCommand struct{}

func (this *UpdateCommand) Execute(args []string) error {
	if err := builds.LoadConfig(); err != nil {
		return err
	}
	if len(args) == 0 {
		return errors.New("not specify operation, available operation [core|tun2socks|geodata|subscribe]").WithPrefix("update").WithPathObj(*this)
	}
	if len(args) > 1 {
		return errors.New("too many arguments").WithPrefix("update").WithPathObj(*this)
	}
	switch args[0] {
	case "core":
		log.HandleInfo("update: updating core")
		if err := updateCore(); err != nil {
			return err
		}
		log.HandleInfo("update: update success")
	case "tun2socks":
		log.HandleInfo("update: updating tun2socks")
		if err := updateTun2socks(); err != nil {
			return err
		}
		log.HandleInfo("update: update success")
	case "geodata":
		log.HandleInfo("update: updating geodata")
		if err := updateGeodata(); err != nil {
			return err
		}
		log.HandleInfo("update: update success")
	case "subscribe":
		log.HandleInfo("update: updating subscribe")
		if err := updateSubscribe(); err != nil {
			return err
		}
		log.HandleInfo("update: update success")
	default:
		return errors.New("unknown operation " + args[0] + ", available operation [core|tun2socks|geodata|subscribe]").WithPrefix("update").WithPathObj(*this)
	}
	return nil
}

// updateCore update core, support xray, singbox
func updateCore() error {
	if runtime.GOARCH != "arm64" {
		return errors.New("this feature only support arm64 device").WithPrefix("update")
	}
	serviceRunFlag := false
	if err := os.MkdirAll(builds.Config.XrayHelper.DataDir, 0644); err != nil {
		return errors.New("create run dir failed, ", err).WithPrefix("update")
	}
	if err := os.MkdirAll(path.Dir(builds.Config.XrayHelper.CorePath), 0644); err != nil {
		return errors.New("create core path dir failed, ", err).WithPrefix("update")
	}
	switch builds.Config.XrayHelper.CoreType {
	case "xray":
		xrayZipPath := path.Join(builds.Config.XrayHelper.DataDir, "xray.zip")
		if err := common.DownloadFile(xrayZipPath, xrayCoreDownloadUrl); err != nil {
			return err
		}
		// update core need stop core service first
		if len(getServicePid()) > 0 {
			log.HandleInfo("update: detect core is running, stop it")
			stopService()
			serviceRunFlag = true
			_ = os.Remove(builds.Config.XrayHelper.CorePath)
		}
		zipReader, err := zip.OpenReader(xrayZipPath)
		if err != nil {
			return errors.New("open xray.zip failed, ", err).WithPrefix("update")
		}
		defer func(zipReader *zip.ReadCloser) {
			_ = zipReader.Close()
			_ = os.Remove(xrayZipPath)
		}(zipReader)
		for _, file := range zipReader.File {
			if file.Name == "xray" {
				fileReader, err := file.Open()
				if err != nil {
					return errors.New("cannot get file reader "+file.Name+", ", err).WithPrefix("update")
				}
				saveFile, err := os.OpenFile(builds.Config.XrayHelper.CorePath, os.O_WRONLY|os.O_CREATE|os.O_SYNC|os.O_TRUNC, 0755)
				if err != nil {
					return errors.New("cannot open file "+builds.Config.XrayHelper.CorePath+", ", err).WithPrefix("update")
				}
				_, err = io.Copy(saveFile, fileReader)
				if err != nil {
					return errors.New("save file "+builds.Config.XrayHelper.CorePath+" failed, ", err).WithPrefix("update")
				}
				_ = saveFile.Close()
				_ = fileReader.Close()
				break
			}
		}
	case "sing-box":
		singboxDownloadUrl, err := getSingboxDownloadUrl()
		if err != nil {
			return err
		}
		singboxGzipPath := path.Join(builds.Config.XrayHelper.DataDir, "sing-box.tar.gz")
		if err := common.DownloadFile(singboxGzipPath, singboxDownloadUrl); err != nil {
			return err
		}
		// update core need stop core service first
		if len(getServicePid()) > 0 {
			log.HandleInfo("update: detect core is running, stop it")
			stopService()
			serviceRunFlag = true
			_ = os.Remove(builds.Config.XrayHelper.CorePath)
		}
		singboxGzip, err := os.Open(singboxGzipPath)
		if err != nil {
			return errors.New("open gzip file failed, ", err).WithPrefix("update")
		}
		defer func(singboxGzip *os.File) {
			_ = singboxGzip.Close()
			_ = os.Remove(singboxGzipPath)
		}(singboxGzip)
		gzipReader, err := gzip.NewReader(singboxGzip)
		if err != nil {
			return errors.New("open gzip file failed, ", err).WithPrefix("update")
		}
		defer func(gzipReader *gzip.Reader) {
			_ = gzipReader.Close()
		}(gzipReader)
		tarReader := tar.NewReader(gzipReader)
		for {
			fileHeader, err := tarReader.Next()
			if err != nil {
				if err == io.EOF {
					return errors.New("cannot find sing-box binary").WithPrefix("update")
				}
				continue
			}
			if filepath.Base(fileHeader.Name) == "sing-box" {
				saveFile, err := os.OpenFile(builds.Config.XrayHelper.CorePath, os.O_WRONLY|os.O_CREATE|os.O_SYNC|os.O_TRUNC, 0755)
				_, err = io.Copy(saveFile, tarReader)
				if err != nil {
					return errors.New("save file "+builds.Config.XrayHelper.CorePath+" failed, ", err).WithPrefix("update")
				}
				_ = saveFile.Close()
				break
			}
		}
	default:
		return errors.New("unknown core type " + builds.Config.XrayHelper.CoreType).WithPrefix("update")
	}
	if serviceRunFlag {
		log.HandleInfo("update: starting core with new version")
		_ = startService()
	}
	return nil
}

// updateTun2socks update tun2socks
func updateTun2socks() error {
	if runtime.GOARCH != "arm64" {
		return errors.New("this feature only support arm64 device").WithPrefix("update")
	}
	savePath := path.Join(path.Dir(builds.Config.XrayHelper.CorePath), "tun2socks")
	if err := common.DownloadFile(savePath, tun2socksDownloadUrl); err != nil {
		return err
	}
	return nil
}

// updateGeodata update geodata
func updateGeodata() error {
	if err := os.MkdirAll(builds.Config.XrayHelper.DataDir, 0644); err != nil {
		return errors.New("create DataDir failed, ", err).WithPrefix("update")
	}
	if err := common.DownloadFile(path.Join(builds.Config.XrayHelper.DataDir, "geoip.dat"), geoipDownloadUrl); err != nil {
		return err
	}
	if err := common.DownloadFile(path.Join(builds.Config.XrayHelper.DataDir, "geosite.dat"), geositeDownloadUrl); err != nil {
		return err
	}
	return nil
}

// updateSubscribe update subscribe
func updateSubscribe() error {
	if err := os.MkdirAll(builds.Config.XrayHelper.DataDir, 0644); err != nil {
		return errors.New("create DataDir failed, ", err).WithPrefix("update")
	}
	builder := strings.Builder{}
	for _, subUrl := range builds.Config.XrayHelper.SubList {
		rawData, err := common.GetRawData(subUrl)
		if err != nil {
			log.HandleError(err)
			continue
		}
		subData, err := common.DecodeBase64(string(rawData))
		if err != nil {
			log.HandleError(err)
			continue
		}
		builder.WriteString(strings.TrimSpace(subData) + "\n")
	}
	if builder.Len() > 0 {
		if err := os.WriteFile(path.Join(builds.Config.XrayHelper.DataDir, "sub.txt"), []byte(builder.String()), 0644); err != nil {
			return errors.New("write subscribe file failed, ", err).WithPrefix("update")
		}
	}
	return nil
}

// getSingboxDownloadUrl use github api to get singbox download url
func getSingboxDownloadUrl() (string, error) {
	rawData, err := common.GetRawData(singboxUrl)
	if err != nil {
		return "", err
	}
	var jsonValue interface{}
	err = json.Unmarshal(rawData, &jsonValue)
	if err != nil {
		return "", errors.New("unmarshal github json failed, ", err).WithPrefix("update")
	}
	// assert json to map
	jsonMap, ok := jsonValue.(map[string]interface{})
	if !ok {
		return "", errors.New("assert github json to map failed").WithPrefix("update")
	}
	assets, ok := jsonMap["assets"]
	if !ok {
		return "", errors.New("cannot find assets ").WithPrefix("update")
	}
	// assert assets
	assetsMap, ok := assets.([]interface{})
	if !ok {
		return "", errors.New("assert assets to []interface failed").WithPrefix("update")
	}
	for _, asset := range assetsMap {
		assetMap, ok := asset.(map[string]interface{})
		if !ok {
			continue
		}
		name, ok := assetMap["name"].(string)
		if !ok {
			continue
		}
		if strings.Contains(name, "android-arm64.tar.gz") {
			downloadUrl, ok := assetMap["browser_download_url"].(string)
			if !ok {
				return "", errors.New("assert browser_download_url to string failed").WithPrefix("update")
			}
			return downloadUrl, nil
		}
	}
	return "", errors.New("cannot get get singbox download url").WithPrefix("update")
}
