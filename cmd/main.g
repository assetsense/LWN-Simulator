package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
)

type C2Config struct {
	C2ServerREST            string `json:"c2serverREST"`
	C2ServerWS              string `json:"c2serverWS"`
	Username                string `json:"username"`
	Password                string `json:"password"`
	CreateDevicesLWN        bool   `json:"createDevicesLWN"`
	JoinDelay               int    `json:"joinDelay"`
	DataPathS               string `json:"dataPathS"`
	DataPathL               string `json:"dataPathL"`
	SendInterval            int    `json:"sendInterval"`
	AckTimeout              int    `json:"ackTimeout"`
	RxDelay                 int    `json:"rxDelay"`
	RXDurationOpen          int    `json:"rxDurationOpen"`
	DataRate                int    `json:"dataRate"`
	ConfigDirName           string `json:"configDirname"`
	MgDeviceId              string `json:"mgDeviceId"`
	MgPasscode              string `json:"mgPasscode"`
	CreateDevicesChirpstack bool   `json:"createDevicesChirpstack"`
	ChirpstackServer        string `json:"chirpstackServer"`
	ApiToken                string `json:"apiToken"`
	ApplicationId           string `json:"applicationId"`
	ProfileId               string `json:"profileId"`
	TenantId                string `json:"tenantId"`
}

func main() {
	config := OpenC2Json()
	if true {
		typ := "fft_x/"
		filesLFFTX, err := os.ReadDir(config.DataPathL + typ)
		if err != nil {
			log.Fatal(err)
		}
		totalFilesLFFTX := len(filesLFFTX)

		for i := 0; i < totalFilesLFFTX; i++ {
			dataPath := config.DataPathL + typ + string(filesLFFTX[0].Name()[0]) + strconv.Itoa(i) + ".bin"
			fmt.Println(dataPath)
			file, err := os.Open(dataPath)
			if err != nil {
				fmt.Println("Error opening file:", err)
				// return ""
			}
			defer file.Close()

			// Read binary data into a buffer
			buffer := make([]byte, 128)
			_, err = file.Read(buffer)
			if err != nil {
				fmt.Println("Error reading binary data:", err)
				// return ""
			}
			data := hex.EncodeToString(buffer)
			fmt.Println(data)
		}
	}
}

func OpenC2Json() C2Config {
	//open c2.json file
	path := "c2.json"

	//c2.json properties can be accessed by config.<property_name>
	config := C2Config{}

	c2Data, err := os.ReadFile(path)
	if err != nil {
		fmt.Println("Error opening c2.json file:", err)
	}

	err = json.Unmarshal(c2Data, &config)
	if err != nil {
		fmt.Println("Error decoding c2.json file:", err)
	}

	return config
}
