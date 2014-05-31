package telize

import (
	//"fmt"
	"encoding/json"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
)

type ResultType int

const (
	RES_SUCCESS ResultType = iota
	RES_ERROR
)

type GetGeoResult struct {
	GeoInfo *GeoIpJson
	Error   *TelizeError
	Type    ResultType
}

type GeoIpJson struct {
	Ip             string
	Country_code   string
	Country_code3  string
	Country        string
	Region_code    string
	Region         string
	City           string
	Postal_code    string
	Continent_code string
	Latitude       float32
	Longitude      float32
	Dma_code       string
	Area_code      string
	Asn            string
	Isp            string
	Timezone       string
}

type TelizeError struct {
	Message string
	Code    int
}

type TelizeRequest struct {
	IP net.IP
}

func (r *TelizeRequest) GetGeo() (*GetGeoResult, error) {
	result := new(GetGeoResult)
	(*result).GeoInfo = new(GeoIpJson)
	(*result).Error = new(TelizeError)
	result.Type = RES_SUCCESS

	requestStrings := []string{"http://www.telize.com/geoip/", r.IP.String()}
	requestString := strings.Join(requestStrings, "")
	resp, err := http.Get(requestString)
	if err != nil {
		return result, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	err = json.Unmarshal(body, result.GeoInfo)
	if err != nil {
		return result, err
	}
	err = json.Unmarshal(body, result.Error)
	if err != nil {
		return result, err
	}

	//determine result type
	if result.Error.Code != 0 {
		result.Type = RES_ERROR
	}
	//process request
	return result, nil
}
