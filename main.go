package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"time"

	"periph.io/x/conn/v3/driver/driverreg"
	"periph.io/x/conn/v3/i2c"
	"periph.io/x/conn/v3/i2c/i2creg"
	"periph.io/x/host/v3"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
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

func initClient() (influxdb2.Client, error) {
	token, found := os.LookupEnv("INFLUXDB_TOKEN")
	if !found {
		return nil, fmt.Errorf("INFLUXDB_TOKEN not set")
	}
	url, found := os.LookupEnv("INFLUXDB_URL")
	if !found {
		url = "http://localhost:8086"
	}
	return influxdb2.NewClient(url, token), nil
}

func initInfo() (InfluxDBInfo, error) {
	org, found := os.LookupEnv("INFLUXDB_ORG")
	if !found {
		org = "lemolatoon"
	}
	bucket, found := os.LookupEnv("INFLUXDB_BUCKET")
	if !found {
		bucket = "sensor-home"
	}
	return InfluxDBInfo{Org: org, Bucket: bucket}, nil
}

type InfluxDBInfo struct {
	Org    string
	Bucket string
}

func initLocation() *time.Location {
	loc, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		log.Printf("Failed to laod Asia/Tokyo timezone: %v", err)
		loc = time.FixedZone("Asia/Tokyo", 9*60*60) // fallback
	}
	return loc
}

func send(client influxdb2.Client, info InfluxDBInfo, loc *time.Location, result *AHT21BResult) {
	writeAPI := client.WriteAPIBlocking(info.Org, info.Bucket)
	tags := map[string]string{
		"sensor": "AHT21B",
	}
	fields := map[string]interface{}{
		"temperature": result.Temperature,
		"humidity":    result.Humidity,
	}
	point := write.NewPoint("sensor_data", tags, fields, time.Now().In(loc))

	if err := writeAPI.WritePoint(context.Background(), point); err != nil {
		log.Printf("Error writing point: %v", err)
	}
}

func initSleepDuration() time.Duration {
	durationStr, found := os.LookupEnv("SLEEP_DURATION_SECONDS")
	if !found {
		durationStr = "60"
	}
	duration, err := strconv.Atoi(durationStr)
	if err != nil {
		log.Printf("Invalid SLEEP_DURATION_SECONDS value: %v, defaulting to 60 seconds", err)
		duration = 60
	}
	if duration <= 0 {
		log.Printf("SLEEP_DURATION_SECONDS must be positive, defaulting to 60 seconds")
		duration = 60
	}
	return time.Duration(duration) * time.Second
}

func main() {
	dev, closer, err := initDevice()
	if err != nil {
		log.Fatal(err)
	}
	defer closer.Close()

	loc := initLocation()
	info, err := initInfo()
	if err != nil {
		log.Fatal(err)
	}
	client, err := initClient()
	if err != nil {
		log.Fatal(err)
	}

	sleepDuration := initSleepDuration()

	for {
		result, err := read(dev)
		if err != nil {
			log.Println("Error: ", err)
		} else {
			fmt.Printf("Temperature: %.2f °C\n", result.Temperature)
			fmt.Printf("Humidity: %.2f %%RH\n", result.Humidity)
		}
		go send(client, info, loc, result)
		time.Sleep(sleepDuration)
	}
}
