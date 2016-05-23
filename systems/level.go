package systems

import (
	"fmt"
	"io"
	"io/ioutil"

	"engo.io/engo"
	"gopkg.in/yaml.v2"
)

type Wave []WaveComponent

type WaveComponent struct {
	Name   string
	Amount int
}

type Level struct {
	Cities []engo.Point
	Waves  []Wave
}

type LevelResource struct {
	Level *Level
	url   string
}

func (l LevelResource) URL() string {
	return l.url
}

type LevelLoader struct {
	levels map[string]LevelResource
}

// Load processes the data stream and parses it as a level file
func (i *LevelLoader) Load(url string, data io.Reader) error {
	b, err := ioutil.ReadAll(data)
	if err != nil {
		return err
	}

	level := new(Level)
	err = yaml.Unmarshal(b, &level)
	if err != nil {
		return err
	}

	i.levels[url] = LevelResource{level, url}
	return nil
}

// Load removes the preloaded audio file from the cache
func (l *LevelLoader) Unload(url string) error {
	delete(l.levels, url)
	return nil
}

// Resource retrieves the preloaded audio file, passed as a `LevelResource`
func (l *LevelLoader) Resource(url string) (engo.Resource, error) {
	texture, ok := l.levels[url]
	if !ok {
		return nil, fmt.Errorf("resource not loaded by `FileLoader`: %q", url)
	}

	return texture, nil
}

func init() {
	engo.Files.Register(".yaml", &LevelLoader{levels: make(map[string]LevelResource)})
}
