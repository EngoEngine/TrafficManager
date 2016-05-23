package systems

import (
	"engo.io/ecs"
	"engo.io/engo"
)

const (
	dispatchButton = "dispatch"
)

type waveEntityCity struct {
	*ecs.BasicEntity
	*CityComponent
}

type WaveSystem struct {
	world  *ecs.World
	cities []waveEntityCity

	waves     []Wave
	waveIndex int
}

func (w *WaveSystem) SetWaves(waves []Wave) {
	w.waveIndex = 0
	w.waves = waves
}

func (w *WaveSystem) AddCity(basic *ecs.BasicEntity, city *CityComponent) {
	w.cities = append(w.cities, waveEntityCity{basic, city})
}

func (w *WaveSystem) Remove(basic ecs.BasicEntity) {
	delete := -1
	for index, e := range w.cities {
		if e.BasicEntity.ID() == basic.ID() {
			delete = index
			break
		}
	}
	if delete >= 0 {
		w.cities = append(w.cities[:delete], w.cities[delete+1:]...)
	}
}

func (w *WaveSystem) New(*ecs.World) {
	engo.Input.RegisterButton(dispatchButton, engo.Space)
}

func (w *WaveSystem) Update(dt float32) {
	if engo.Input.Button(dispatchButton).JustPressed() {
		if w.waveIndex == len(w.waves) {
			return // we're done :)
		}

		wave := w.waves[w.waveIndex]
		for _, fromCity := range wave {
			w.cities[fromCity.From].Enqueue(fromCity.Vehicles)
		}
		w.waveIndex++
	}
}
