package tlnet

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/donnie4w/json4g"
	"github.com/donnie4w/simplelog/logging"
)

type Ip struct {
	Province string
	City     string
	Area     string
	Allarea  string //地址全称
	IpAddr   string // ip地址
	Isp      string //运营商
}

func (this *Ip) Parse(ip string) {
	this.IpAddr = ip
	urladdr := "https://ip.useragentinfo.com/json?ip="
	resp, err := http.Get(urladdr + this.IpAddr)
	if err == nil {
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		logging.Debug(err)
		if err == nil {
			json4g, e := json4g.LoadByString(string(body))
			logging.Debug(e)
			if e == nil {
				this.Province = json4g.GetNodeByName("province").ValueString
				this.City = json4g.GetNodeByName("city").ValueString
				this.Area = json4g.GetNodeByName("area").ValueString
				this.Isp = json4g.GetNodeByName("isp").ValueString
				this.Allarea = fmt.Sprint(this.Province, this.City, this.Area)
			}
		}
	}
}
