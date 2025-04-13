package main

import (
	"fmt"
	"io"
	"log"
	"time"

	"periph.io/x/conn/v3/driver/driverreg"
	"periph.io/x/conn/v3/i2c"
	"periph.io/x/conn/v3/i2c/i2creg"
	"periph.io/x/host/v3"
)

const AHT21BAddr = 0x38

type AHT21BResult struct {
	Temperature float32
	Humidity    float32
}

func initDevice() (*i2c.Dev, io.Closer, error) {
	if _, err := host.Init(); err != nil {
		return nil, nil, fmt.Errorf("failed to initialize periph: %v", err)
	}

	if _, err := driverreg.Init(); err != nil {
		return nil, nil, fmt.Errorf("failed to initialize driverreg: %v", err)
	}

	bus, err := i2creg.Open("")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open I2C bus: %v", err)
	}

	dev := &i2c.Dev{
		Addr: AHT21BAddr,
		Bus:  bus,
	}

	return dev, bus, nil
}

func read(dev *i2c.Dev) (*AHT21BResult, error) {
	write := []byte{0xAC, 0x33, 0x00}
	if err := dev.Tx(write, nil); err != nil {
		return nil, fmt.Errorf("failed to write command: %v", err)
	}

	time.Sleep(80 * time.Millisecond)

	read := make([]byte, 6)
	if err := dev.Tx(nil, read); err != nil {
		return nil, fmt.Errorf("failed to read data: %v", err)
	}

	if read[0]&0x80 != 0 {
		return nil, fmt.Errorf("sensor is busy")
	}

	rawHumi := (uint32(read[1]) << 12) | (uint32(read[2]) << 4) | (uint32(read[3]) >> 4)
	rawTemp := (uint32(read[3]&0x0F) << 16) | (uint32(read[4]) << 8) | uint32(read[5])

	humidity := float32(rawHumi) * 100 / 1048576.0
	temperature := float32(rawTemp)*200/1048576.0 - 50

	return &AHT21BResult{
		Temperature: temperature,
		Humidity:    humidity,
	}, nil
}

func main() {
	dev, closer, err := initDevice()
	if err != nil {
		log.Fatal(err)
	}
	defer closer.Close()
	for {
		result, err := read(dev)
		if err != nil {
			log.Println("Error: ", err)
		} else {
			fmt.Printf("Temperature: %.2f °C\n", result.Temperature)
			fmt.Printf("Humidity: %.2f %%RH\n", result.Humidity)
		}
	}
}
