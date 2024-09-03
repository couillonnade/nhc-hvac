# nbc-hvac

This service sets the FAN speed of Niko Home Control HVAC automatically.

Parameters :

    IP and Port of the Niko Controller
    "broker": "mqtts://ipaddress:8884"

    Credentials from Niko Controller
    "username": "hobby"
    "password": "certificatfromniko"

    Uuid of the Thermostat
    "HVAC-TH-Uuid": "guid-from-hvac-thermostat-device"

    Delta Range for fan speed in celcius degrees
    "SmallTempDelta": 0.5
    "ModerateTempDelta": 3

    Hysteresis in minutes to limit updates
    "Hysteresis": 10
