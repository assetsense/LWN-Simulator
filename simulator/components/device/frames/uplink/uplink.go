package uplink

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/arslab/lwnsimulator/simulator/components/device/features/adr"
	mac "github.com/arslab/lwnsimulator/simulator/components/device/macCommands"
	"github.com/arslab/lwnsimulator/simulator/util"
	"github.com/brocaar/lorawan"
)

type InfoUplink struct {
	DwellTime     lorawan.DwellTime `json:"-"`
	ClassB        bool              `json:"-"`
	FCnt          uint32            `json:"fcnt"`
	FOpts         []lorawan.Payload `json:"-"`
	FPort         *uint8            `json:"fport"`
	ADR           adr.ADRInfo       `json:"-"`
	AckMacCommand mac.AckMacCommand `json:"-"` //to create new Uplink
}

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

func (up *InfoUplink) GetFrame(mtype lorawan.MType, payload lorawan.DataPayload,
	devAddr lorawan.DevAddr, AppSKey, NwkSKey [16]byte, ack bool) ([]byte, error) {

	FOpts := up.loadFOpts()

	phy := lorawan.PHYPayload{
		MHDR: lorawan.MHDR{
			MType: mtype,
			Major: lorawan.LoRaWANR1,
		},
		MACPayload: &lorawan.MACPayload{
			FHDR: lorawan.FHDR{
				DevAddr: devAddr,
				FCtrl: lorawan.FCtrl{
					ADR:       up.ADR.ADR,
					ADRACKReq: up.ADR.ADRACKReq,
					ACK:       ack,
					ClassB:    up.ClassB,
				},
				FCnt:  up.FCnt,
				FOpts: FOpts,
			},
			FPort: up.FPort,
			FRMPayload: []lorawan.Payload{
				&payload,
			},
		},
	}

	bytes, err := encryptFrame(phy, AppSKey, NwkSKey)
	if err != nil {
		return []byte{}, err
	}

	up.FCnt = (up.FCnt + 1) % util.MAXFCNTGAP
	up.ADR.ADRACKCnt++

	return bytes, nil

}

func (up *InfoUplink) loadFOpts() []lorawan.Payload {

	FOpts := up.AckMacCommand.GetAll()
	if len(up.FOpts) > 0 {

		if len(up.FOpts)+len(FOpts) < 15 {
			FOpts = append(FOpts, up.FOpts...)
			up.FOpts = up.FOpts[:0] //reset
		} else {
			FOpts = append(FOpts, up.FOpts[:15-len(FOpts)]...)
			up.FOpts = up.FOpts[15-len(FOpts):] //reset
		}

	}

	return FOpts
}

func encryptFrame(phy lorawan.PHYPayload, AppSKey, NwkSKey [16]byte) ([]byte, error) {

	if err := phy.EncryptFRMPayload(AppSKey); err != nil {
		return []byte{}, err
	}

	if err := phy.SetUplinkDataMIC(lorawan.LoRaWAN1_0, 0, 0, 0, NwkSKey, lorawan.AES128Key{}); err != nil {
		return []byte{}, err
	}

	bytes, err := phy.MarshalBinary()
	if err != nil {
		return []byte{}, err
	}

	return bytes, nil
}

func (up *InfoUplink) IsTherePingSlotInfoReq() bool {

	for _, cmd := range up.FOpts {

		cid, _, err := mac.ParseMACCommand(cmd, true)
		if err != nil {
			return false
		}

		if cid == lorawan.PingSlotInfoReq {
			return true
		}
	}

	return false

}

var config C2Config = OpenC2Json()

func Fragmentation(deviceType int, dataType int, axisId int, size int, payloadDiscard lorawan.Payload) []lorawan.DataPayload {

	// fmt.Println(payloadDiscard.MarshalBinary())
	var FRMPayload []lorawan.DataPayload

	payloads := ReadDataSample(deviceType, dataType, axisId, config)

	for _, payloadBytes := range payloads {
		// fmt.Println(payloadBytes)
		// payloadBytes := []byte(payload)

		if size == 0 {
			return FRMPayload
		}

		nFrame := len(payloadBytes) / size

		for i := 0; i <= nFrame; i++ {

			var data lorawan.DataPayload

			offset := i * size

			if i != nFrame {
				data.Bytes = payloadBytes[offset : offset+size]
			} else {
				data.Bytes = payloadBytes[offset:len(payloadBytes)]
			}

			FRMPayload = append(FRMPayload, data)
		}
	}
	return FRMPayload
}

func Truncate(size int, payload lorawan.Payload) lorawan.DataPayload {
	var FRMPayload lorawan.DataPayload

	payloadBytes, _ := payload.MarshalBinary()

	if len(payloadBytes) > size {
		FRMPayload.Bytes = payloadBytes[:size]
	} else {
		FRMPayload.Bytes = payloadBytes
	}

	return FRMPayload
}

//*******************************JSON**************************************/

func (up *InfoUplink) MarshalJSON() ([]byte, error) {

	type Alias InfoUplink

	return json.Marshal(&struct {
		FPort uint8 `json:"fport"`
		*Alias
	}{

		FPort: *up.FPort,
		Alias: (*Alias)(up),
	})

}

func (up *InfoUplink) UnmarshalJSON(data []byte) error {

	type Alias InfoUplink

	aux := &struct {
		FPort uint8 `json:"fport"`
		*Alias
	}{
		Alias: (*Alias)(up),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	up.FPort = &aux.FPort

	return nil
}

func GetDataSample(dataType string, config C2Config) [][]byte {
	var samples [][]byte
	if dataType == "s" {
		//s-type
		dataPath := config.DataPathS + "s.bin"
		file, err := os.Open(dataPath)
		if err != nil {
			fmt.Println("Error opening file:", err)
			return nil
		}
		defer file.Close()

		fileInfo, err := file.Stat()
		if err != nil {
			fmt.Println("Error:", err)
			return nil
		}
		fileSize := fileInfo.Size()

		// Read binary data into a buffer
		buffer := make([]byte, fileSize)
		_, err = file.Read(buffer)
		if err != nil {
			fmt.Println("Error reading binary data:", err)
			return nil
		}
		// data := hex.EncodeToString(buffer)
		// samples = append(samples, data)
		samples = append(samples, buffer)
		return samples
	}
	//l-type
	filesL, err := os.ReadDir(config.DataPathL + dataType)
	if err != nil {
		log.Fatal(err)
	}
	totalFilesL := len(filesL)

	for i := 0; i < totalFilesL; i++ {
		dataPath := config.DataPathL + dataType + string(filesL[0].Name()[0]) + strconv.Itoa(i) + ".bin"
		file, err := os.Open(dataPath)
		if err != nil {
			fmt.Println("Error opening file:", err)
			return nil
		}
		defer file.Close()

		fileInfo, err := file.Stat()
		if err != nil {
			fmt.Println("Error:", err)
			return nil
		}
		fileSize := fileInfo.Size()
		// fmt.Println(fileSize)

		// Read binary data into a buffer
		buffer := make([]byte, fileSize)
		_, err = file.Read(buffer)
		if err != nil {
			fmt.Println("Error reading binary data:", err)
			return nil
		}
		// data := hex.EncodeToString(buffer)
		// samples = append(samples, data)
		samples = append(samples, buffer)
	}

	return samples
}

func ReadDataSample(deviceType int, dataType int, axisId int, config C2Config) [][]byte {
	if deviceType == 6199 {
		//S-Type
		return GetDataSample("s", config)

	} else if deviceType == 6165 {
		//L-Type

		if dataType == 6163 {
			//psd
			if axisId == 6166 {
				//x-axis
				return GetDataSample("psd_raw_x/", config)

			} else if axisId == 6167 {
				//y-axis
				return GetDataSample("psd_raw_y/", config)

			} else if axisId == 6168 {
				//z-axis
				return GetDataSample("psd_raw_z/", config)

			} else if axisId == 6169 {
				//tri-axial
				return GetDataSample("psd_raw_triaxis/", config)

			}

		} else if dataType == 6164 {
			//fft
			if axisId == 6166 {
				//x-axis
				return GetDataSample("fft_x/", config)

			} else if axisId == 6167 {
				//y-axis
				return GetDataSample("fft_y/", config)

			} else if axisId == 6168 {
				//z-axis
				return GetDataSample("fft_z/", config)

			} else if axisId == 6169 {
				//tri-axial
				return GetDataSample("fft_triaxis/", config)
			}

		} else if dataType == 6179 {
			//fft-raw
			if axisId == 6166 {
				//x-axis
				return GetDataSample("fft_raw_x/", config)

			} else if axisId == 6167 {
				//y-axis
				return GetDataSample("fft_raw_y/", config)

			} else if axisId == 6168 {
				//z-axis
				return GetDataSample("fft_raw_z/", config)

			} else if axisId == 6169 {
				//tri-axial
				return GetDataSample("fft_raw_triaxis/", config)
			}

		}

	}
	return nil
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
