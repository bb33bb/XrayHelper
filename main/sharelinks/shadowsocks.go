package sharelinks

import (
	"XrayHelper/main/errors"
	"XrayHelper/main/serial"
	"XrayHelper/main/utils"
	"net/url"
	"strconv"
	"strings"
)

const ssPrefix = "ss://"

type Shadowsocks struct {
	nodeName string
	address  string
	port     uint16
	method   string
	password string
}

func (this *Shadowsocks) GetNodeInfo() string {
	return serial.Concat("Node Name: ", this.nodeName, ", Type: Shadowsocks, Address: ", this.address, ", Port: ", this.port, ", Method: ", this.method, ", Password: ", this.password)
}

func (this *Shadowsocks) ToOutoundJsonWithTag(tag string) string {
	//TODO
	return ""
}

func NewShadowsocksShareLink(ssUrl string) (ShareLink, error) {
	ss := new(Shadowsocks)
	if !strings.HasPrefix(ssUrl, ssPrefix) {
		return nil, errors.New(ssUrl + " is not a shadowsocks share link").WithPrefix("shadowsocks")
	}
	ssUrl = strings.TrimPrefix(ssUrl, ssPrefix)
	nodeAndName := strings.Split(ssUrl, "#")
	nodeName, err := url.QueryUnescape(nodeAndName[1])
	if err != nil {
		return nil, errors.New("unescape node name failed, ", err).WithPrefix("shadowsocks")
	}
	ss.nodeName = nodeName
	infoAndServer := strings.Split(nodeAndName[0], "@")
	addressAndPort := strings.Split(infoAndServer[1], ":")
	ss.address = addressAndPort[0]
	port, err := strconv.Atoi(addressAndPort[1])
	if err != nil {
		return nil, errors.New("convert node port failed, ", err).WithPrefix("shadowsocks")
	}
	ss.port = uint16(port)
	info, err := utils.DecodeBase64(infoAndServer[0])
	if err != nil {
		return nil, err
	}
	methodAndPassword := strings.Split(info, ":")
	ss.method = methodAndPassword[0]
	ss.password = methodAndPassword[1]
	return ss, nil
}
