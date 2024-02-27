package util

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"

	"github.com/arslab/lwnsimulator/models"
)

func GetPath() (string, error) {

	homepath, errorPath := os.Getwd()
	if errorPath != nil {
		return "", errorPath
	}

	path := homepath + "/" + GetConfigDirname()
	err := CreateConfigDir(path)
	if err != nil {
		return "", err
	}

	return path, nil
}

func GetConfigDirname() string {

	info, err := models.GetConfigFile("c2.json")
	if err != nil {
		log.Fatal(err)
	}

	return info.ConfigDirname

}

func CreateConfigDir(path string) error {
	return os.MkdirAll(path, os.ModePerm)
}

func CreateConfigFiles() {

	var path [3]string

	pathDir, err := GetPath()
	if err != nil {
		log.Fatal(err)
	}

	path[0] = pathDir + "/simulator.json"
	path[1] = pathDir + "/gateways.json"
	path[2] = pathDir + "/devices.json"

	for i, _ := range path {

		data := "{}"
		if i == 0 {
			data = `{
				"nextIDDev": 0,
				"nextIDGw": 1,
				"bridgeAddress": "localhost:1700"
			}`
		} else if i == 1 {
			data = `{
				"0": {
					"id": 0,
					"info": {
						"macAddress": "c8fd50b13d204e38",
						"keepAlive": 30,
						"active": true,
						"typeGateway": false,
						"name": "test-gw",
						"location": {
							"latitude": 0,
							"longitude": 0,
							"altitude": 0
						},
						"ip": "",
						"port": ""
					}
				}
			}`
		}

		_, err = os.Create(path[i])
		if err != nil {
			log.Fatal(err)
		}

		WriteConfigFile(path[i], []byte(data))
	}

}

func RecoverConfigFile(path string, v interface{}) error {

	if _, err := os.Stat(path); os.IsNotExist(err) {
		CreateConfigFiles()
	}

	fileBytes, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	return json.Unmarshal(fileBytes, &v)

}

func WriteConfigFile(path string, data []byte) error {

	if _, err := os.Stat(path); os.IsNotExist(err) {

		_, err = os.Create(path)
		if err != nil {
			log.Fatal(err)
		}

	}

	return ioutil.WriteFile(path, data, os.ModePerm)
}
