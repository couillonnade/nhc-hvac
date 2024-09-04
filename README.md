# nbc-hvac

This minimalist service sets the FAN speed of Niko Home Control HVAC automatically.
You must compile for the target platform and update the config.json file (see below).

It can be launch on a terminal or there is a basic script available to install it as a service on linux. You need to place the config file and binary in a folder and set chchmod +x ./install.sh to launch the script. 


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

## Usefull Go Commands

Cheat sheet for Go and RevPi

### Modules / Packages

SUMDB might have 30mn delay when updating, so a way around it if versions is not availaible is to set it off:
go clean --modcache
GOSUMDB=off go get -u github.com/package...

### Cross-Compile

**Linux 64-bits:** env GOOS=linux GOARCH=amd64 go build main.go

**Raspberry Pi 3+/4:** env GOOS=linux GOARCH=arm GOARM=7 go build -o ~/Dev/github/nhc/bin

Get CPU Information:

cat /proc/cpuinfo
[(See Also)](https://github.com/goreleaser/goreleaser/issues/36)

### RevPi Core 3+

sudo teamviewer-iot-agent stop daemon -> not enough (comes back after restart)
even if not listed on here: service --status-all

remove the package:
teamviewer-revpi
