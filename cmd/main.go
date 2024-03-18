package main

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
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

type ResponseBatch struct {
	MsgType        string      `json:"msg_type"`
	FullImport     bool        `json:"fullImport"`
	FinalBatch     bool        `json:"finalBatch"`
	DataSize       int         `json:"dataSize"`
	Sequence       int         `json:"sequence"`
	MgDevice       interface{} `json:"mg_device"`
	BondedDevices  interface{} `json:"bonded_devices"`
	DeviceProfiles interface{} `json:"deviceProfiles"`
	Timestamp      int64       `json:"timestamp"`
}

var devLat = 0
var pgLat = 500

func main() {

	simulatorRepository := repo.NewSimulatorRepository()
	simulatorController := cnt.NewSimulatorController(simulatorRepository)
	simulatorController.GetIstance()

	log.Println("LWN Simulator is online...")
	log.Println("Press enter to exit anytime!")

	config := OpenC2Json()

	if config.CreateDevices {
		// AddDevicesToSimulatorREST(simulatorController, config)
		AddDevicesToSimulatorWS(simulatorController, config)
	}

	// <-done
	simulatorController.Run()

	//the main goroutine finishes before other sub goroutines, due to which the program exits before
	//finishing the sub goroutines, the user input blocks the main go routine to finish
	BlockMainRoutine()
}

func AddDevicesToSimulatorWS(simulatorController cnt.SimulatorController, config C2Config) {
	//fetch all the devices from c2 WS as json string
	GetDevicesFromC2WS(simulatorController, config)

	// fmt.Println(config.MgDeviceId)
	// filePath := "batch3.txt"

	// fileContent, err := ioutil.ReadFile(filePath)
	// if err != nil {
	// 	log.Fatal("Error reading file:", err)
	// 	return
	// }

	// var responseBatch ResponseBatch
	// err = json.Unmarshal(fileContent, &responseBatch)

	// if err != nil {
	// 	log.Println("Unmarshal error of file:", err)
	// 	return
	// }

	// AddDevicesToSimulatorWSHelper(simulatorController, config, responseBatch)

}

func AddDevicesToSimulatorWSHelper(simulatorController cnt.SimulatorController, config C2Config, responseBatch ResponseBatch) int {

	devicesReceived := 0

	// Unmarshal the JSON data into an interface{}
	var bondedDevices = responseBatch.BondedDevices

	// Assert the interface{} to a slice of interfaces ([]interface{})
	devices, ok := bondedDevices.([]interface{})
	if !ok {
		fmt.Println("Error: Bonded Devices JSON data is not a slice of interfaces")
		return 0
	}

	//opening the datasamples directory
	filesS, err := os.ReadDir(config.DataPathS)
	if err != nil {
		log.Fatal(err)
	}

	filesL, err := os.ReadDir(config.DataPathL)
	if err != nil {
		log.Fatal(err)
	}

	totalFilesS := len(filesS)
	totalFilesL := len(filesL)
	iS := 0
	iL := 0

	// Iterate over devices
	for _, device := range devices {
		devicesReceived += 1
		deviceMap, ok := device.(map[string]interface{})
		if !ok {
			fmt.Println("Error: Invalid device format")
			continue
		}

		// Access specific properties
		deviceType, _ := deviceMap["type"].(float64)
		// deviceID, _ := deviceMap["id"].(float64)
		deviceEui, _ := deviceMap["code"].(string)
		deviceName, _ := deviceMap["name"].(string)
		appKey, _ := deviceMap["key"].(string)
		axisId, _ := deviceMap["axis"].(float64)
		profileId, _ := deviceMap["profileId"].(float64)

		profileMap := getProfileMap(profileId, responseBatch)

		deviceSupportOTAA := profileMap["deviceSupportOTAA"].(bool)
		deviceSupportClassB := profileMap["deviceSupportClassB"].(bool)
		deviceSupportClassC := profileMap["deviceSupportClassC"].(bool)
		region := getRegionId(profileMap["deviceRegion"].(string))

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
			pg := getPgJson(deviceEui, deviceName, pgLat)

			// Convert to JSON string
			jsonData, err := json.MarshalIndent(pg, "", "    ")
			if err != nil {
				fmt.Println("Error:", err)
				return 0
			}

			var pgObj gateway.Gateway
			errr := json.Unmarshal([]byte(string(jsonData)), &pgObj)
			if errr != nil {
				fmt.Println("Error in pgObj:", errr)
				return 0
			}

			// log.Println(deviceName)
			_, id, _ := simulatorController.AddGateway(&pgObj)
			if id == 0 {

			}
			pgLat = pgLat + 1000
			continue

		} else {
			//ignoring other devices like MGs...
			continue
		}

		// Convert binary data to a hexadecimal string
		payloadData = ReadDataSample(dataPath)

		// Create an instance of DeviceJSON
		device := getDeviceJson(deviceEui, deviceName, appKey, devLat, payloadData, region, deviceSupportOTAA, deviceSupportClassB, deviceSupportClassC, config)

		// Convert to JSON string
		jsonData, err := json.MarshalIndent(device, "", "    ")
		if err != nil {
			fmt.Println("Error in marshal device:", err)
			return 0
		}

		var deviceObj dev.Device
		errr := json.Unmarshal([]byte(string(jsonData)), &deviceObj)
		if errr != nil {
			fmt.Println("Error in unmarshal deviceobj:", errr)
			return 0
		}

		_, id, err := simulatorController.AddDevice(&deviceObj)
		if id == 0 {

		}
		devLat = devLat + 25
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

func GetDevicesFromC2WS(simulatorController cnt.SimulatorController, config C2Config) {

	// interrupt := make(chan os.Signal, 1)
	// signal.Notify(interrupt, os.Interrupt)

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

	// Send the request message
	message := fmt.Sprintf(`{"msg_type": "req_bonded_devices", "device": "%s", "ls": 0}`, config.MgDeviceId)
	err = c.WriteMessage(websocket.TextMessage, []byte(message))
	if err != nil {
		log.Fatal("Write error:", err)
	}

	log.Println("Waiting for devices from WS... (Restart if no devices received)")

	dataSize := 0
	devicesReceived := 0
	prevSequence := 0

	// Handle incoming messages from the WebSocket server
	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			log.Println("Read error:", err)
			return
		}
		// fmt.Println(string(message))
		var responseBatch ResponseBatch
		err = json.Unmarshal(message, &responseBatch)
		if err != nil {
			log.Println("Unmarshal error:", err)
			continue
		}

		if responseBatch.MsgType == "resp_bonded_devices" {

			if responseBatch.Sequence != prevSequence+1 {
				log.Println("Batch ", prevSequence+1, " lost! Please restart.")
				return
			}
			prevSequence = responseBatch.Sequence

			if dataSize == 0 {
				dataSize = responseBatch.DataSize
			}

			// fmt.Println(bondedDevices)
			devicesReceived += AddDevicesToSimulatorWSHelper(simulatorController, config, responseBatch)

			if responseBatch.FinalBatch {
				log.Println("Devices received: ", devicesReceived, " | Total: ", dataSize)
				//send close message
				errr := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
				if errr != nil {
					log.Println("Write close error:", errr)
				}

				c.Close()
				return
			}
		} else {
			log.Println("Batching message structure is incorrect!")
			return
		}
	}

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

func BlockMainRoutine() {
	var userInput string
	_, err := fmt.Scanln(&userInput)
	if err != nil {
		fmt.Println("Simulator Exited", err)
		return
	}
}

func hashString(s string) int {
	h := fnv.New32a()
	h.Write([]byte(s))
	return int(h.Sum32())
}

func getPgJson(deviceEui string, deviceName string, pgLat int) PgJson {
	pg := PgJson{
		ID: hashString(deviceEui),
		PGInfo: PGInfo{
			MacAddress:  deviceEui,
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

	return pg
}

func getDeviceJson(deviceEui string, deviceName string, appKey string, devLat int, payloadData string, region int, deviceSupportOTAA bool, deviceSupportClassB bool, deviceSupportClassC bool, config C2Config) DeviceJSON {
	device := DeviceJSON{
		ID: hashString(deviceEui),
		Info: Info{
			Name:    deviceName,
			DevEUI:  deviceEui,
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
				Region:            region,
				SendInterval:      config.SendInterval,
				AckTimeout:        config.AckTimeout,
				Range:             10000,
				DisableFCntDown:   true,
				SupportedOTAA:     deviceSupportOTAA,
				SupportedADR:      false,
				SupportedFragment: true,
				SupportedClassB:   deviceSupportClassB,
				SupportedClassC:   deviceSupportClassC,
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
	return device
}

func ReadDataSample(path string) string {
	// Open the binary data file
	file, err := os.Open(path)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return ""
	}
	defer file.Close()

	// Read binary data into a buffer
	buffer := make([]byte, 128)
	_, err = file.Read(buffer)
	if err != nil {
		fmt.Println("Error reading binary data:", err)
		return ""
	}
	return hex.EncodeToString(buffer)
}

func getProfileMap(profileId float64, responseBatch ResponseBatch) map[string]interface{} {

	var deviceProfiles = responseBatch.DeviceProfiles
	var returnProfile map[string]interface{}
	// Assert the interface{} to a slice of interfaces ([]interface{})
	profiles, ok := deviceProfiles.([]interface{})
	if !ok {
		fmt.Println("Error: Device Proifles JSON data is not a slice of interfaces")
		return returnProfile
	}

	for _, profile := range profiles {
		ProfileMap, _ := profile.(map[string]interface{})
		returnProfile = ProfileMap

		id, _ := ProfileMap["id"].(float64)
		if id == profileId {
			return ProfileMap
		}
	}
	return returnProfile
}

func getRegionId(region string) int {
	if region == "EU868" {
		return 1
	}
	if region == "US915" {
		return 2
	}
	if region == "CN779" {
		return 3
	}
	if region == "EU433" {
		return 4
	}
	if region == "AU915" {
		return 5
	}
	if region == "CN470" {
		return 6
	}
	if region == "AS923" {
		return 7
	}
	if region == "KR920" {
		return 8
	}
	if region == "IN865" {
		return 9
	}
	if region == "RU864" {
		return 10
	}
	return 1
}
