const fs = require('fs');
const c2 = require('./c2.json');

const username = c2.username;
const password = c2.password;

const authString = `${username}:${password}`;
const encodedAuth = btoa(authString);

var devices = {};

async function getDevicesFromC2() {

    await fetch(c2.c2server, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            'Authorization': `Basic ${encodedAuth}`, // Include your Basic Authentication credentials here
        },
        body: '{}',
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
            console.error(error);
            return;
        });
}

async function addDevicesToSimulator() {

    for (let i = 0; i < devices.length; i++) {
        var dev = devices[i];

        if (!dev.hasOwnProperty("applicationKey")) {
            continue;
        }
        var binaryData;
        //change to device type
        if (dev.deviceType.id == 6199) {
            binaryData = fs.readFileSync(c2.dataPathS).toString('hex');
        } else if (dev.deviceType.id = 6165) {
            binaryData = fs.readFileSync(c2.dataPathL).toString('hex');
        } else {
            binaryData = "";
        }
        var deviceJson = {
            "id": dev.deviceCode,
            "info": {
                "name": dev.deviceName,
                "devEUI": dev.deviceCode + "",
                "appKey": dev.applicationKey,
                "devAddr": "00000000",
                "nwkSKey": "00000000000000000000000000000000",
                "appSKey": "00000000000000000000000000000000",
                "location": {
                    "latitude": 0,
                    "longitude": 0,
                    "altitude": 0
                },
                "status": {
                    "mtype": "ConfirmedDataUp",
                    "payload": binaryData,
                    "active": true,
                    "infoUplink": {
                        "fport": 1,
                        "fcnt": 1
                    },
                    "fcntDown": 0
                },
                "configuration": {
                    "region": 1,
                    "sendInterval": c2.sendInterval,
                    "ackTimeout": c2.ackTimeout,
                    "range": 10000,
                    "disableFCntDown": true,
                    "supportedOtaa": true,
                    "supportedADR": false,
                    "supportedFragment": true,
                    "supportedClassB": false,
                    "supportedClassC": false,
                    "dataRate": 0,
                    "rx1DROffset": 0,
                    "nbRetransmission": 1
                },
                "rxs": [
                    {
                        "delay": c2.rxDelay,
                        "durationOpen": c2.rxDurationOpen,
                        "channel": {
                            "active": false,
                            "enableUplink": false,
                            "freqUplink": 0,
                            "freqDownlink": 0,
                            "minDR": 0,
                            "maxDR": 0
                        },
                        "dataRate": c2.dataRate
                    },
                    {
                        "delay": c2.rxDelay,
                        "durationOpen": c2.rxDurationOpen,
                        "channel": {
                            "active": true,
                            "enableUplink": false,
                            "freqUplink": 0,
                            "freqDownlink": 869525000,
                            "minDR": 0,
                            "maxDR": 0
                        },
                        "dataRate": c2.dataRate
                    }
                ]
            }
        };


        await fetch(c2.simulatorServer + "api/add-device", {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(deviceJson),
        })
            .then(response => {
                return response.json();
            })
            .then(data => {
                console.log(dev.deviceName);
                if (data.code == 0) {
                    console.log("Device added successfully!");
                } else {
                    console.log(data.status);
                }
            })
            .catch(error => {
                console.log('Error: simulator webserver is not running');
            });
    };



}

async function startSimulator() {
    await fetch(c2.simulatorServer + "api/start")
        .then(response => {
            if (!response.ok) {
                throw new Error(`HTTP error! Status: ${response.status}`);
            }
            return response.json();
        })
        .then(data => {
            console.log("Simulator has been started!");
        })
        .catch(error => {
            console.error('Error: simulator webserver is not running');
        });
}

async function main() {

    if (c2.createDevices == true) {
        await getDevicesFromC2();
        await addDevicesToSimulator();
    }
    //await startSimulator();
}

main();
