const grpc = require("@grpc/grpc-js");
const c2 = require('./c2.json')
const { DeviceServiceClient} = require("@chirpstack/chirpstack-api/api/device_grpc_pb");
const {CreateDeviceRequest, CreateDeviceKeysRequest ,DeviceQueueItem } = require("@chirpstack/chirpstack-api/api/device_pb");
const {Device , DeviceKeys} = require("@chirpstack/chirpstack-api/api/device_pb");

// This must point to the ChirpStack API interface.
const server = c2.chirpstackServer;

var devices = {};

async function getDevicesFromC2() {
    const apiUrl = c2.apiUrl;
    const username = c2.username;
    const password = c2.password;
    const postData = "{}";
    const authString = `${username}:${password}`;
    const encodedAuth = btoa(authString);
    await fetch(apiUrl, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            'Authorization': `Basic ${encodedAuth}`, // Include your Basic Authentication credentials here
        },
        body: postData,
    })
        .then(response => {
            if (!response.ok) {
                throw new Error('Network response was not ok');
            }
            return response.json();
        })
        .then(data => {
            devices = data.Device;
        })
        .catch(error => {
            console.error('Error:', error);
        });
}

async function main() {
    await getDevicesFromC2();
    const appId = "5e373f1b-eea5-4ecf-8165-20099817724a";
    const profileId = "e02422c6-f46d-42c1-8c63-756fda1d3c62";
    const apiToken = "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJhdWQiOiJjaGlycHN0YWNrIiwiaXNzIjoiY2hpcnBzdGFjayIsInN1YiI6IjYwOGNjMWZkLWRkMjgtNDEzMy1iOTVkLWJjODNkY2I4ZjA3MSIsInR5cCI6ImtleSJ9.IkOwMfjAwyKeM5r1w2gwicOVRqVTmc-l6na5fORjx54";
    await addingDevice(server,apiToken,devices,appId,profileId);
}
main()
async function addingDevice(server,apiToken,deviceList,applicationId,deviceProfileId){

    //check if device is not an array
    if (!Array.isArray(deviceList)) {
        deviceList = [deviceList];
    }

    for (const dev of deviceList) {
        if (!dev.hasOwnProperty("applicationKey")) {
            continue;
        }
        if (!dev.hasOwnProperty("networkKey")) {
            continue;
        }
        try {
            //setting up the device
            const req = new CreateDeviceRequest();
            const device = new Device();

            device.setDevEui(dev.deviceCode);
            device.setName(dev.deviceName);
            device.setApplicationId(applicationId);
            device.setDeviceProfileId(deviceProfileId);
            device.setDescription("Registering device via API");
            device.setIsDisabled(false);
            device.setSkipFcntCheck(false);
            device.setIsDisabled(false);
            req.setDevice(device);

            //setting up the device key
            const keysReq = new CreateDeviceKeysRequest();
            const keys = new DeviceKeys();

            keys.setDevEui(dev.deviceCode);
            keys.setAppKey(dev.applicationKey);
            keys.setNwkKey(dev.applicationKey);
            keysReq.setDeviceKeys(keys);

            
            //adding the device and the device key
            const channel = await new DeviceServiceClient(server, grpc.credentials.createInsecure());
            const metadata = await new grpc.Metadata();
            await metadata.set("authorization", "Bearer " + apiToken);
            await channel.create(req, metadata, async(err, resp) => {
                console.log(dev.deviceName);
                if (err !== null) {
                    console.log("Device is already added");
                    return;
                } else {
                    console.log("Device added successfully!");
                }
                await channel.createKeys(keysReq, metadata, async (err, resp) => {
                    if (err !== null) {
                        console.log("CreateKeys error occured");
                        return;
                    } else {
                        console.log("Device keys created!");
                    }
                });
            });
        } catch (e) {
            console.log("Error at adding device key", e);
            return response.status(500).send(`An error occurred: ${e}`);
        }
    }
    return true;
}