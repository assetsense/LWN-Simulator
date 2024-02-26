package main

import (
	// "fmt"
	"io"
	"log"
	"os/exec"

	// "sync"
	// "time"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	// "strings"
	cnt "github.com/arslab/lwnsimulator/controllers"
	// "github.com/arslab/lwnsimulator/models"
	repo "github.com/arslab/lwnsimulator/repositories"
	// ws "github.com/arslab/lwnsimulator/webserver"

	dev "github.com/arslab/lwnsimulator/simulator/components/device"

	"encoding/hex"
	"os"
	"strconv"
)

type DeviceType struct {
	ID            int    `json:"id"`
	Category      int    `json:"category"`
	Code          string `json:"code"`
	Default       bool   `json:"default"`
	Description   string `json:"description"`
	Name          string `json:"name"`
	Position      int    `json:"position"`
	Purpose       string `json:"purpose"`
	SystemDefined bool   `json:"systemDefined"`
}

// DeviceJSON represents the structure you want to create
type DeviceJSON struct {
	ID   int  `json:"id"`
	Info Info `json:"info"`
}

// Info represents the "info" part of the structure
type Info struct {
	Name          string        `json:"name"`
	DevEUI        string        `json:"devEUI"`
	AppKey        string        `json:"appKey"`
	DevAddr       string        `json:"devAddr"`
	NwkSKey       string        `json:"nwkSKey"`
	AppSKey       string        `json:"appSKey"`
	Location      Location      `json:"location"`
	Status        Status        `json:"status"`
	Configuration Configuration `json:"configuration"`
	RXs           []RX          `json:"rxs"`
}

// Location represents the "location" part of the structure
type Location struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Altitude  float64 `json:"altitude"`
}

// Status represents the "status" part of the structure
type Status struct {
	MType      string `json:"mtype"`
	Payload    string `json:"payload"`
	Active     bool   `json:"active"`
	InfoUplink struct {
		FPort int `json:"fport"`
		FCnt  int `json:"fcnt"`
	} `json:"infoUplink"`
	FCntDown int `json:"fcntDown"`
}

// Configuration represents the "configuration" part of the structure
type Configuration struct {
	Region            int  `json:"region"`
	SendInterval      int  `json:"sendInterval"`
	AckTimeout        int  `json:"ackTimeout"`
	Range             int  `json:"range"`
	DisableFCntDown   bool `json:"disableFCntDown"`
	SupportedOTAA     bool `json:"supportedOtaa"`
	SupportedADR      bool `json:"supportedADR"`
	SupportedFragment bool `json:"supportedFragment"`
	SupportedClassB   bool `json:"supportedClassB"`
	SupportedClassC   bool `json:"supportedClassC"`
	DataRate          int  `json:"dataRate"`
	RX1DROffset       int  `json:"rx1DROffset"`
	NbRetransmission  int  `json:"nbRetransmission"`
}

// RX represents the "rxs" part of the structure
type RX struct {
	Delay        int     `json:"delay"`
	DurationOpen int     `json:"durationOpen"`
	Channel      Channel `json:"channel"`
	DataRate     int     `json:"dataRate"`
}

// Channel represents the "channel" part of the structure within RX
type Channel struct {
	Active       bool `json:"active"`
	EnableUplink bool `json:"enableUplink"`
	FreqUplink   int  `json:"freqUplink"`
	FreqDownlink int  `json:"freqDownlink"`
	MinDR        int  `json:"minDR"`
	MaxDR        int  `json:"maxDR"`
}

type C2Config struct {
	SimulatorServer  string `json:"simulatorServer"`
	ChirpstackServer string `json:"chirpstackServer"`
	C2Server         string `json:"c2server"`
	Username         string `json:"username"`
	Password         string `json:"password"`
	CreateDevices    bool   `json:"createDevices"`
	JoinDelay        int    `json:"joinDelay"`
	DataPathS        string `json:"dataPathS"`
	DataPathL        string `json:"dataPathL"`
	SendInterval     int    `json:"sendInterval"`
	AckTimeout       int    `json:"ackTimeout"`
	RxDelay          int    `json:"rxDelay"`
	RXDurationOpen   int    `json:"rxDurationOpen"`
	DataRate         int    `json:"dataRate"`
	ConfigDirName    string `json:"configDirname"`
}

func RunCommand() (string, error) {
	cmd := exec.Command("node", "cmd/add-devices-simulator.js")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("error running command: %v", err)
	}

	return string(output), nil
}

func getDevicesFromC2() string {
	path := "c2.json"

	config := C2Config{}

	c2Data, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Println("Error opening file:", err)
	}

	err = json.Unmarshal(c2Data, &config)
	if err != nil {
		fmt.Println("Error decoding JSON:", err)
	}
	apiURL := config.C2Server
	username := config.Username
	password := config.Password
	postData := "{}"

	authString := fmt.Sprintf("%s:%s", username, password)
	encodedAuth := base64.StdEncoding.EncodeToString([]byte(authString))

	req, err := http.NewRequest("POST", apiURL, bytes.NewBufferString(postData))
	if err != nil {
		fmt.Println("Error creating request:", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Basic "+encodedAuth)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error making request:", err)
	}
	defer resp.Body.Close()

	result, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error decoding response:", err)
	}

	if result == nil {
		fmt.Println("error: Device not found")
	}

	return string(result)
}

func main() {

	// var info *models.ServerConfig
	// var err error

	simulatorRepository := repo.NewSimulatorRepository()
	simulatorController := cnt.NewSimulatorController(simulatorRepository)
	simulatorController.GetIstance()

	log.Println("LWN Simulator is online...")
	fmt.Println("Press enter to exit!")

	// Open the JSON file
	path := "c2.json"

	config := C2Config{}

	c2Data, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}

	err = json.Unmarshal(c2Data, &config)
	if err != nil {
		fmt.Println("Error decoding JSON:", err)
		return
	}

	if config.CreateDevices == true {
		jsonData := getDevicesFromC2()

		var data map[string]interface{}
		errr := json.Unmarshal([]byte(jsonData), &data)
		if errr != nil {
			fmt.Println("Error: Api Url is not valid")
			return
		}

		// Access the "Device" array
		devices, ok := data["Device"].([]interface{})
		if !ok {
			fmt.Println("Error: Credentials is invalid | Device array not found in JSON")
			return
		}

		// file, err := os.Open(path + "/c2.json")
		// if err != nil {
		// 	fmt.Println("Error opening file:", err)
		// 	return
		// }
		// defer file.Close()

		// Decode the JSON file into a struct
		// var config C2Config
		// decoder := json.NewDecoder(file)
		// err = decoder.Decode(&config)
		// if err != nil {
		// 	fmt.Println("Error decoding JSON:", err)
		// 	return
		// }

		// Iterate over devices
		var ltypeCnt int = 0
		for _, device := range devices {

			deviceMap, ok := device.(map[string]interface{})
			if !ok {
				fmt.Println("Error: Invalid device format")
				continue
			}

			deviceType, ok := deviceMap["deviceType"].(map[string]interface{})
			if !ok {
				fmt.Println("Error: device type not found")
				continue
			}
			deviceId, _ := deviceType["id"].(float64)
			// fmt.Println("\n\n\n", deviceId)
			var dataPath string
			var payloadData string
			if deviceId == 6199 {
				dataPath = config.DataPathS
			} else if deviceId == 6165 {
				dataPath = config.DataPathL + strconv.Itoa(ltypeCnt) + ".bin"
				ltypeCnt = ltypeCnt + 1
				if ltypeCnt > 120 {
					ltypeCnt = 0
				}
			} else {
				// dataPath = path + config.DataPathS
				continue
			}
			// Open the binary file
			file, err := os.Open(dataPath)
			if err != nil {
				fmt.Println("Error opening file:", err)
				return
			}
			defer file.Close()

			// Read binary data into a buffer
			buffer := make([]byte, 128)
			_, err = file.Read(buffer)
			if err != nil {
				fmt.Println("Error reading binary data:", err)
				return
			}

			// Convert binary data to a hexadecimal string
			payloadData = hex.EncodeToString(buffer)

			// Access specific properties
			deviceID, _ := deviceMap["id"].(int)
			deviceEui, _ := deviceMap["deviceCode"].(string)
			deviceName, _ := deviceMap["deviceName"].(string)
			appKey, _ := deviceMap["applicationKey"].(string)

			//this is implemented because there is an issue in the rest service (when the device id doesn't
			//contain any character then it treats as an integer.
			var deviceEuiint float64
			var Euiint = false
			var deviceEuistring string
			if deviceEui == "" {
				Euiint = true
				deviceEuiint = deviceMap["deviceCode"].(float64)
			}
			if Euiint == true {
				deviceEuistring = strconv.FormatFloat(deviceEuiint, 'f', -1, 64)
			} else {
				deviceEuistring = deviceEui
			}

			// Create an instance of DeviceJSON
			device := DeviceJSON{
				ID: deviceID,
				Info: Info{
					Name:    deviceName,
					DevEUI:  deviceEuistring,
					AppKey:  appKey,
					DevAddr: "00000000",
					NwkSKey: "00000000000000000000000000000000",
					AppSKey: "00000000000000000000000000000000",
					Location: Location{
						Latitude:  0,
						Longitude: 0,
						Altitude:  0,
					},
					Status: Status{
						MType:   "ConfirmedDataUp",
						Payload: payloadData,
						Active:  true,
						InfoUplink: struct {
							FPort int `json:"fport"`
							FCnt  int `json:"fcnt"`
						}{
							FPort: 1,
							FCnt:  1,
						},
						FCntDown: 0,
					},
					Configuration: Configuration{
						Region:            1,
						SendInterval:      config.SendInterval,
						AckTimeout:        config.AckTimeout,
						Range:             10000,
						DisableFCntDown:   true,
						SupportedOTAA:     true,
						SupportedADR:      false,
						SupportedFragment: true,
						SupportedClassB:   false,
						SupportedClassC:   false,
						DataRate:          0,
						RX1DROffset:       0,
						NbRetransmission:  1,
					},
					RXs: []RX{
						{
							Delay:        config.RxDelay,
							DurationOpen: config.RXDurationOpen,
							Channel: Channel{
								Active:       false,
								EnableUplink: false,
								FreqUplink:   0,
								FreqDownlink: 0,
								MinDR:        0,
								MaxDR:        0,
							},
							DataRate: config.DataRate,
						},
						{
							Delay:        config.RxDelay,
							DurationOpen: config.RXDurationOpen,
							Channel: Channel{
								Active:       true,
								EnableUplink: false,
								FreqUplink:   0,
								FreqDownlink: 869525000,
								MinDR:        0,
								MaxDR:        0,
							},
							DataRate: config.DataRate,
						},
					},
				},
			}

			// Convert to JSON string
			jsonData, err := json.MarshalIndent(device, "", "    ")
			if err != nil {
				fmt.Println("Error:", err)
				return
			}

			// fmt.Println(string(jsonData))
			var deviceObj dev.Device
			errr := json.Unmarshal([]byte(string(jsonData)), &deviceObj)
			if errr != nil {
				// fmt.Println(&deviceObj)
				fmt.Println("Error:", errr)
				return
			}

			code, id, err := simulatorController.AddDevice(&deviceObj)
			log.Println(deviceName)
			if code == 0 || id == 0 {
				log.Println("added successfully")
			}

		}
	}

	simulatorController.Run()
	var userInput string
	_, errr := fmt.Scanln(&userInput)
	if errr != nil {
		fmt.Println("Simulator Exited", err)
		return
	}

	// info, err = models.GetConfigFile("config.json")
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// WebServer := ws.NewWebServer(info, simulatorController)
	// //WebServer.Run()

	// var wg sync.WaitGroup

	// wg.Add(1)

	// go func() {
	// 	defer wg.Done()
	// 	WebServer.Run()
	// }()

	// time.Sleep(10 * time.Second)
	// result, err := RunCommand()
	// if err != nil {
	// 	fmt.Println("Error:", err)
	// 	return
	// }

	// fmt.Println("Command Output:")
	// fmt.Println(result)

}
