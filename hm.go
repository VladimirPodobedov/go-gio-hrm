package main

import (
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/widget"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/permisions"
	"tinygo.org/x/bluetooth"
	"strconv"	
	"os"
)

var (
	adapter = bluetooth.DefaultAdapter

	heartRateServiceUUID        = bluetooth.ServiceUUIDHeartRate
	heartRateCharacteristicUUID = bluetooth.CharacteristicUUIDHeartRateMeasurement
)

func main() {
	permisions.RequestPermision("< PERMISION NAME >")
	a := app.New()
	w := a.NewWindow("Heart Monitor")

	log := binding.NewString()
	log.Set("starting...")
	log_label := widget.NewLabelWithData(log)
	
	go func(){
	
		// Enable BLE interface.
		log.Set("Enable BLE interface...")
		err := adapter.Enable()
		if err != nil {
			log.Set("failed to " + "enable BLE stack" + ": " + err.Error())
			return
		}
	
		ch := make(chan bluetooth.ScanResult, 1)
	
		// Scan device
		log.Set("Scanning...")
		err1 := adapter.Scan(func(adapter *bluetooth.Adapter, result bluetooth.ScanResult) {
			log.Set("found device:" + result.Address.String() + " " + result.LocalName())
			if result.Address.String() == connectAddress() {
				adapter.StopScan()
				ch <- result
			}
		})
		if err1 != nil {
			log.Set("failed to " + "scanning" + ": " + err.Error())
			return
		}				
	
		// Connect to device 
		log.Set("Connecting...")
		var device bluetooth.Device
		select {
		case result := <-ch:
			device, err = adapter.Connect(result.Address, bluetooth.ConnectionParams{})
			if err != nil {
				log.Set(err.Error())
				return
			}
			log.Set("connected to " + result.Address.String())
		}
	
		// Discovering services
		log.Set("Discovering services..")
		srvcs, err := device.DiscoverServices([]bluetooth.UUID{heartRateServiceUUID})
		if err != nil {
			log.Set("failed to " + "discover services" + ": " + err.Error())
			return
		}		
		if len(srvcs) == 0 {
			log.Set("could not find heart rate service")
			return
		}
	
		srvc := srvcs[0]

		log.Set("found service" + srvc.UUID().String())

		// Discovering characteristics
		log.Set("Discovering characteristics..")
		chars, err := srvc.DiscoverCharacteristics([]bluetooth.UUID{heartRateCharacteristicUUID})
		if err != nil {
			log.Set("could not find heart rate characteristics")
			return
		}

		if len(chars) == 0 {
			log.Set("could not find heart rate characteristic")
			return
		}
		
		char := chars[0]
		log.Set("found characteristic" + char.UUID().String())
		char.EnableNotifications(func(buf []byte) {
			log.Set("data: " + strconv.FormatUint(uint64(uint8(buf[1])),10))
		})
		select {}
	}()
	
	w.SetContent(log_label)
	w.ShowAndRun()
}

func must(action string, err error) {
	if err != nil {
		panic("failed to " + action + ": " + err.Error())
	}
}

func connectAddress() string {
	if len(os.Args) < 2 {
		println("usage: heartrate-monitor [address]")
		os.Exit(1)
	}

	// look for device with specific name
	address := os.Args[1]

	return address
}

