const c2 = require('./c2.json')
const apiUrl = c2.apiUrl;
const username = c2.username;
const password = c2.password;
const postData = "{}";
const authString = `${username}:${password}`;
const encodedAuth = btoa(authString);

var devices = {};

async function getDevicesFromC2() {
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

async function addDevicesToSimulator() {

    for (let i = 0; i < devices.length; i++) {
        var dev = devices[i];

        if (!dev.hasOwnProperty("applicationKey")) {
            continue;
        }
        console.log(dev.deviceName);
        var deviceJson = {
            "id": dev.deviceCode,
            "info": {
                "name": dev.deviceName,
                "devEUI": dev.deviceCode + "",
                "ecn": dev.ecn,
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
                    "payload": "",
                    "active": true,
                    "infoUplink": {
                        "fport": 1,
                        "fcnt": 1
                    },
                    "fcntDown": 0
                },
                "configuration": {
                    "region": 1,
                    "sendInterval": 30,
                    "ackTimeout": 10,
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
                        "delay": 15000,
                        "durationOpen": 30000,
                        "channel": {
                            "active": false,
                            "enableUplink": false,
                            "freqUplink": 0,
                            "freqDownlink": 0,
                            "minDR": 0,
                            "maxDR": 0
                        },
                        "dataRate": 0
                    },
                    {
                        "delay": 15000,
                        "durationOpen": 30000,
                        "channel": {
                            "active": true,
                            "enableUplink": false,
                            "freqUplink": 0,
                            "freqDownlink": 869525000,
                            "minDR": 0,
                            "maxDR": 0
                        },
                        "dataRate": 0
                    }
                ]
            }
        };


        await fetch("http://localhost:8000/api/add-device", {
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
                if (data.code == 0) {
                    console.log("Device added successfully!");
                } else {
                    console.log(data.status);
                }
            })
            .catch(error => {
                console.log('Error: webserver is not running');
            });
    };



}

async function startSimulator() {
    const apiUrl = "http://localhost:8000/api/start";

    fetch(apiUrl)
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
            console.error('Error during fetch:', error.message);
        });
}

async function main() {
    console.log("main");
    await getDevicesFromC2();
    await addDevicesToSimulator();
    await startSimulator();
}


main();
