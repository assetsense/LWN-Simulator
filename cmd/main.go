package main

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"

	cnt "github.com/arslab/lwnsimulator/controllers"
	repo "github.com/arslab/lwnsimulator/repositories"
	dev "github.com/arslab/lwnsimulator/simulator/components/device"
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
	C2Server       string `json:"c2server"`
	Username       string `json:"username"`
	Password       string `json:"password"`
	CreateDevices  bool   `json:"createDevices"`
	JoinDelay      int    `json:"joinDelay"`
	DataPathS      string `json:"dataPathS"`
	DataPathL      string `json:"dataPathL"`
	SendInterval   int    `json:"sendInterval"`
	AckTimeout     int    `json:"ackTimeout"`
	RxDelay        int    `json:"rxDelay"`
	RXDurationOpen int    `json:"rxDurationOpen"`
	DataRate       int    `json:"dataRate"`
	ConfigDirName  string `json:"configDirname"`
}

func getDevicesFromC2() string {

	//opening c2.json file
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
	//c2.json properties can be accessed by config.<property_name>

	apiURL := config.C2Server
	username := config.Username
	password := config.Password
	postData := "{}"

	//creating authentication string
	authString := fmt.Sprintf("%s:%s", username, password)
	encodedAuth := base64.StdEncoding.EncodeToString([]byte(authString))

	//post request to c2 server
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

	//returning the devices as json string
	return string(result)
}

func main() {

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

	//c2.json properties can be accessed by config.<property_name>
	if config.CreateDevices {

		//fetch all the devices from c2 as json string
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
		//opening the datasamples directory

		filesS, err := ioutil.ReadDir(config.DataPathS)
		if err != nil {
			log.Fatal(err)
		}

		filesL, err := ioutil.ReadDir(config.DataPathL)
		if err != nil {
			log.Fatal(err)
		}

		totalFilesS := len(filesS)
		totalFilesL := len(filesL)
		iS := 0
		iL := 0

		// Iterate over devices
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
			var dataPath string
			var payloadData string

			//setting the binary data path as per device type
			if deviceId == 6199 {
				//S-Type
				dataPath = config.DataPathS + filesS[iS].Name()
				iS = iS + 1
				if iS >= totalFilesS {
					iS = 0
				}
			} else if deviceId == 6165 {
				//L-Type
				axis, ok := deviceMap["axis"].(map[string]interface{})
				if !ok {
					fmt.Println("Error: device type not found")
					continue
				}

				axisId, _ := axis["id"].(float64)
				if axisId == 6166 {
					//x-axis

				} else if axisId == 6167 {
					//y-axis

				} else if axisId == 6168 {
					//z-axis

				} else if axisId == 6169 {
					//tri-axial

				}
				dataPath = config.DataPathL + filesL[iL].Name()
				iL = iL + 1
				if iL >= totalFilesL {
					iL = 0
				}
			} else {
				// ignoring other devices like MGs, PGs etc.
				continue
			}
			// Open the binary data file
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
			deviceID, _ := deviceMap["id"].(float64)
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
			if Euiint {
				deviceEuistring = strconv.FormatFloat(deviceEuiint, 'f', -1, 64)
			} else {
				deviceEuistring = deviceEui
			}

			// Create an instance of DeviceJSON
			device := DeviceJSON{
				ID: int(deviceID),
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

			var deviceObj dev.Device
			errr := json.Unmarshal([]byte(string(jsonData)), &deviceObj)
			if errr != nil {
				fmt.Println("Error:", errr)
				return
			}

			log.Println(deviceName)
			code, id, err := simulatorController.AddDevice(&deviceObj)
			if code == 0 || id == 0 {
				log.Println("device added successfully")
			}

		}
	}
	//run the simulator
	simulatorController.Run()

	//the main goroutine finishes before other sub goroutines, due to which the program exits before
	//finishing the sub goroutines, the user input blocks the main go routine to finish
	var userInput string
	_, errr := fmt.Scanln(&userInput)
	if errr != nil {
		fmt.Println("Simulator Exited", err)
		return
	}

}
