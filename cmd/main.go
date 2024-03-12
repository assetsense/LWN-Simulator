package main

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"

	"github.com/gorilla/websocket"

	cnt "github.com/arslab/lwnsimulator/controllers"
	repo "github.com/arslab/lwnsimulator/repositories"
	dev "github.com/arslab/lwnsimulator/simulator/components/device"
	"github.com/arslab/lwnsimulator/simulator/components/gateway"
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

type PgJson struct {
	ID     int    `json:"id"`
	PGInfo PGInfo `json:"info"`
}

type PGInfo struct {
	MacAddress  string   `json:"macAddress"`
	KeepAlive   int      `json:"keepAlive"`
	Active      bool     `json:"active"`
	TypeGateway bool     `json:"typeGateway"`
	Name        string   `json:"name"`
	Location    Location `json:"location"`
	IP          string   `json:"ip"`
	Port        string   `json:"port"`
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
	C2ServerREST   string `json:"c2serverREST"`
	C2ServerWS     string `json:"c2serverWS"`
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
	MgDeviceId     string `json:"mgDeviceId"`
}

type BatchMessage struct {
	MsgType string `json:"msg_type"`
	Msg     string `json:"msg"`
}

type BatchData struct {
	ID           int     `json:"id"`
	LastSyncTime float64 `json:"lastSyncTime"`
	Org          int     `json:"org"`
	CustomerID   int     `json:"customerId"`
	FullImport   bool    `json:"fullImport"`
	FinalBatch   bool    `json:"finalBatch"`
	DataSize     int     `json:"dataSize"`
	Sequence     int     `json:"sequence"`
	Data         string  `json:"data"`
}

func main() {

	simulatorRepository := repo.NewSimulatorRepository()
	simulatorController := cnt.NewSimulatorController(simulatorRepository)
	simulatorController.GetIstance()

	log.Println("LWN Simulator is online...")
	fmt.Println("Press enter to exit!")

	//open c2.json file
	path := "c2.json"

	//c2.json properties can be accessed by config.<property_name>
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

	if config.CreateDevices {
		// AddDevicesToSimulatorREST(simulatorController, config)
		AddDevicesToSimulatorWS(simulatorController, config)
	}

	//run the simulator
	// simulatorController.Run()

	//the main goroutine finishes before other sub goroutines, due to which the program exits before
	//finishing the sub goroutines, the user input blocks the main go routine to finish
	var userInput string
	_, errr := fmt.Scanln(&userInput)
	if errr != nil {
		fmt.Println("Simulator Exited", errr)
		return
	}

}

func GetDevicesFromC2WS(simulatorController cnt.SimulatorController, config C2Config) {

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	apiURL := config.C2ServerWS
	username := config.Username
	password := config.Password

	//creating authentication string
	authString := fmt.Sprintf("%s:%s", username, password)
	encodedAuth := base64.StdEncoding.EncodeToString([]byte(authString))

	// Establish WebSocket connection with Basic Authorization header
	headers := make(http.Header)
	headers.Set("Authorization", "Basic "+encodedAuth)

	// Establish WebSocket connection
	c, _, err := websocket.DefaultDialer.Dial(apiURL, headers)
	if err != nil {
		log.Fatal("Dial error:", err)
	}
	fmt.Println("Connected to WebSocket!")

	// Send the request message
	message := "{\"msg_type\": \"req_bonded_devices\",\"device\":\"" + "30233a09dab4d76a" + "\", \"ls\": 1705591809117}"
	err = c.WriteMessage(websocket.TextMessage, []byte(message))
	if err != nil {
		log.Fatal("Write error:", err)
	}

	fmt.Println("Sent message: " + message)

	done := make(chan struct{})
	dataSize := 0
	devicesReceived := 0

	// Handle incoming messages from the WebSocket server
	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("Read error:", err)
				return
			}
			// fmt.Println(string(message))
			// return
			var batchMessage BatchMessage
			err = json.Unmarshal(message, &batchMessage)

			if err != nil {
				log.Println("Unmarshal error:", err)
				continue
			}

			if batchMessage.MsgType == "batch_out" {
				var batchData BatchData
				err = json.Unmarshal([]byte(batchMessage.Msg), &batchData)
				if err != nil {
					log.Println("Unmarshal batch data error:", err)
					continue
				}
				dataSize = batchData.DataSize
				jsonData := batchData.Data

				// fmt.Println(jsonData)
				devicesReceived += AddDevicesToSimulatorWSHelper(simulatorController, config, jsonData)
				fmt.Println("Number of Devices Received: ", devicesReceived)

				if devicesReceived < dataSize {
					fmt.Println("All devices are not received yet.")
				}

				if batchData.FinalBatch {
					fmt.Println("Final batch received. Stopping reading.")
					//send close message
					errr := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
					if errr != nil {
						log.Println("Write close error:", err)
					}

					// Finally, close the WebSocket connection
					c.Close()
					return
				}
			}
		}
	}()

	// Wait for the goroutine handling incoming messages to finish
	// <-done

}

func AddDevicesToSimulatorWS(simulatorController cnt.SimulatorController, config C2Config) {
	//fetch all the devices from c2 WS as json string
	GetDevicesFromC2WS(simulatorController, config)

	// filePath := "batch1final.txt"

	// // Read the contents of the file
	// fileContent, err := ioutil.ReadFile(filePath)
	// if err != nil {
	// 	log.Fatal("Error reading file:", err)
	// 	return
	// }

	// AddDevicesToSimulatorWSHelper(simulatorController, config, string(fileContent))

}

func AddDevicesToSimulatorWSHelper(simulatorController cnt.SimulatorController, config C2Config, jsonData string) int {
	devicesReceived := 0
	var data map[string]interface{}
	errr := json.Unmarshal([]byte(jsonData), &data)
	if errr != nil {
		fmt.Println("Error:", errr)
		return 0
	}
	// Access the "Data" array
	msgData, _ := data["msg"].(map[string]interface{})
	devices, ok := msgData["data"].([]interface{})
	if !ok {
		fmt.Println("Error: Credentials is invalid | Data array not found in JSON")
		return 0
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

	devLat := 0
	pgLat := 500
	// Iterate over devices
	for _, device := range devices {
		devicesReceived += 1
		deviceMap, ok := device.(map[string]interface{})
		if !ok {
			fmt.Println("Error: Invalid device format")
			continue
		}

		deviceType, ok := deviceMap["deviceType"].(float64)
		if !ok {
			fmt.Println("Error: device type not found")
			continue
		}

		// Access specific properties
		deviceID, _ := deviceMap["id"].(float64)
		deviceEui, _ := deviceMap["deviceCode"].(string)
		deviceName, _ := deviceMap["deviceName"].(string)
		appKey, _ := deviceMap["appKey"].(string)

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

		var dataPath string
		var payloadData string

		//setting the binary data path as per device type
		if deviceType == 6199 {
			//S-Type
			dataPath = config.DataPathS + filesS[iS].Name()
			iS = iS + 1
			if iS >= totalFilesS {
				iS = 0
			}
		} else if deviceType == 6165 {
			//L-Type
			axisId, ok := deviceMap["axis"].(float64)
			if !ok {
				fmt.Println("Error: device type not found")
				continue
			}

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

		} else if deviceType == 6149 {
			device := PgJson{
				ID: int(deviceID),
				PGInfo: PGInfo{
					MacAddress:  deviceEuistring,
					KeepAlive:   30,
					Active:      true,
					TypeGateway: false,
					Name:        deviceName,
					Location: Location{
						Latitude:  float64(pgLat),
						Longitude: 0,
						Altitude:  0,
					},
					IP:   "",
					Port: "",
				},
			}

			// Convert to JSON string
			jsonData, err := json.MarshalIndent(device, "", "    ")
			if err != nil {
				fmt.Println("Error:", err)
				return 0
			}

			var deviceObj gateway.Gateway
			errr := json.Unmarshal([]byte(string(jsonData)), &deviceObj)
			if errr != nil {
				fmt.Println("Error:", errr)
				return 0
			}

			// log.Println(deviceName)
			_, id, _ := simulatorController.AddGateway(&deviceObj)
			if id == 0 {

			}
			pgLat = pgLat + 1000
			continue

		} else {
			//ignoring other devices like MGs...
			continue
		}

		// Open the binary data file
		file, err := os.Open(dataPath)
		if err != nil {
			fmt.Println("Error opening file:", err)
			return 0
		}
		defer file.Close()

		// Read binary data into a buffer
		buffer := make([]byte, 128)
		_, err = file.Read(buffer)
		if err != nil {
			fmt.Println("Error reading binary data:", err)
			return 0
		}

		// Convert binary data to a hexadecimal string
		payloadData = hex.EncodeToString(buffer)

		// deviceEuistring, err = generateRandomEUI()
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
					Latitude:  float64(devLat),
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
					DataRate:          config.DataRate,
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
			return 0
		}

		var deviceObj dev.Device
		errr := json.Unmarshal([]byte(string(jsonData)), &deviceObj)
		if errr != nil {
			fmt.Println("Error:", errr)
			return 0
		}

		// log.Println(deviceName)
		_, id, err := simulatorController.AddDevice(&deviceObj)
		if id == 0 {

		}
		devLat = devLat + 25
		// if code == 0 || id == 0 {
		// 	log.Println("device added successfully")
		// }
	}
	return devicesReceived
}

func AddDevicesToSimulatorREST(simulatorController cnt.SimulatorController, config C2Config) {

	//fetch all the devices from c2 REST as json string
	jsonData := GetDevicesFromC2REST(config)

	var data map[string]interface{}
	errr := json.Unmarshal([]byte(jsonData), &data)
	if errr != nil {
		fmt.Println("Error:", errr)
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

	devLat := 0
	pgLat := 500
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

		// Access specific properties
		deviceId, _ := deviceType["id"].(float64)
		deviceID, _ := deviceMap["id"].(float64)
		deviceEui, _ := deviceMap["deviceCode"].(string)
		deviceName, _ := deviceMap["deviceName"].(string)
		appKey, _ := deviceMap["applicationKey"].(string)
		// gatewayId, _ := deviceMap["gatewayId"].(float64)

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

		} else if deviceId == 6149 {
			device := PgJson{
				ID: int(deviceID),
				PGInfo: PGInfo{
					MacAddress:  deviceEuistring,
					KeepAlive:   30,
					Active:      true,
					TypeGateway: false,
					Name:        deviceName,
					Location: Location{
						Latitude:  float64(pgLat),
						Longitude: 0,
						Altitude:  0,
					},
					IP:   "",
					Port: "",
				},
			}

			// Convert to JSON string
			jsonData, err := json.MarshalIndent(device, "", "    ")
			if err != nil {
				fmt.Println("Error:", err)
				return
			}

			var deviceObj gateway.Gateway
			errr := json.Unmarshal([]byte(string(jsonData)), &deviceObj)
			if errr != nil {
				fmt.Println("Error:", errr)
				return
			}

			// log.Println(deviceName)
			_, id, _ := simulatorController.AddGateway(&deviceObj)
			if id == 0 {

			}
			pgLat = pgLat + 1000
			continue

		} else {
			//ignoring other devices like MGs...
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

		// deviceEuistring, err = generateRandomEUI()
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
					Latitude:  float64(devLat),
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
					DataRate:          config.DataRate,
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

		// log.Println(deviceName)
		_, id, err := simulatorController.AddDevice(&deviceObj)
		if id == 0 {

		}
		devLat = devLat + 25
		// if code == 0 || id == 0 {
		// 	log.Println("device added successfully")
		// }
	}
}

func GetDevicesFromC2REST(config C2Config) string {

	apiURL := config.C2ServerREST
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

func GenerateRandomEUI() (string, error) {
	// Generate a random byte slice of length 8
	randomBytes := make([]byte, 8)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", err
	}

	// Set the two most significant bits (bits 7 and 6) to indicate a globally unique EUI
	randomBytes[0] |= 0xC0 // 0xC0 is 11000000 in binary

	// Convert the byte slice to a hexadecimal string
	eui := hex.EncodeToString(randomBytes)

	return eui, nil
}
