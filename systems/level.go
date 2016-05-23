package systems

import (
	"fmt"
	"io"
	"io/ioutil"

	"engo.io/engo"
	"gopkg.in/yaml.v2"
	"strings"
)

type Vehicles struct {
	Vehicles []Vehicle
}

type Vehicle struct {
	Name         string
	Length       float32
	Minspeed     float32
	Maxspeed     float32
	Brakes       float32
	Acceleration float32
	Cost         int
	Reward       int64
}

type Wave []WaveFromCity

type WaveFromCity struct {
	From     int
	Vehicles []WaveComponent
}

type WaveComponent struct {
	Name   string
	Amount int
	To     CityCategory
}

type LevelCity struct {
	X        float32
	Y        float32
	Category CityCategory
}

type Level struct {
	Cities []LevelCity
	Waves  []Wave
}

type LevelResource struct {
	Level *Level
	url   string
}

func (l LevelResource) URL() string {
	return l.url
}

type VehicleResource struct {
	Vehicles *Vehicles
	url      string
}

func (v VehicleResource) URL() string {
	return v.url
}

type LevelLoader struct {
	levels   map[string]LevelResource
	vehicles map[string]VehicleResource
}

// Load processes the data stream and parses it as a level file
func (i *LevelLoader) Load(url string, data io.Reader) error {
	b, err := ioutil.ReadAll(data)
	if err != nil {
		return err
	}

	switch {
	case strings.HasSuffix(url, "level.yaml"):
		level := new(Level)
		err = yaml.Unmarshal(b, level)
		if err != nil {
			return err
		}

		i.levels[url] = LevelResource{level, url}
	case strings.HasSuffix(url, "vehicles.yaml"):
		vehicles := new(Vehicles)
		err = yaml.Unmarshal(b, vehicles)
		if err != nil {
			return err
		}

		i.vehicles[url] = VehicleResource{vehicles, url}
	}

	return nil
}

// Load removes the preloaded audio file from the cache
func (l *LevelLoader) Unload(url string) error {
	delete(l.levels, url)
	return nil
}

// Resource retrieves the preloaded audio file, passed as a `LevelResource`
func (l *LevelLoader) Resource(url string) (engo.Resource, error) {
	switch {
	case strings.HasSuffix(url, "level.yaml"):
		t, ok := l.levels[url]
		if !ok {
			return nil, fmt.Errorf("resource not loaded by `FileLoader`: %q", url)
		}

		return t, nil

	case strings.HasSuffix(url, "vehicles.yaml"):
		v, ok := l.vehicles[url]
		if !ok {
			return nil, fmt.Errorf("resource not loaded by `FileLoader`: %q", url)
		}

		return v, nil

	default:
		return nil, fmt.Errorf("resource not supported by LevelLoader: %q", url)
	}
}

func init() {
	engo.Files.Register(".yaml", &LevelLoader{
		levels:   make(map[string]LevelResource),
		vehicles: make(map[string]VehicleResource),
	})
}
