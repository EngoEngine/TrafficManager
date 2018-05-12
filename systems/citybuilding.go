package systems

import (
	"engo.io/ecs"
	"engo.io/engo"
	"engo.io/engo/common"

	"fmt"
)

const cityTextureLocation = "textures/city.png"

// CityAssets are the assets for the system
var CityAssets = []string{cityTextureLocation}

type CityMouseTracker struct {
	ecs.BasicEntity
	common.MouseComponent
}

type City struct {
	ecs.BasicEntity
	common.RenderComponent
	common.SpaceComponent
}

type CityBuildingSystem struct {
	world *ecs.World

	mouseTracker CityMouseTracker
}

// New is the initialisation of the System
func (cb *CityBuildingSystem) New(w *ecs.World) {
	cb.world = w
	fmt.Println("CityBuildingSystem was added to the Scene")

	cb.mouseTracker.BasicEntity = ecs.NewBasic()
	cb.mouseTracker.MouseComponent = common.MouseComponent{Track: true}

	engo.Input.RegisterButton("build", engo.KeyQ)

	for _, system := range w.Systems() {
		switch sys := system.(type) {
		case *common.MouseSystem:
			sys.Add(&cb.mouseTracker.BasicEntity, &cb.mouseTracker.MouseComponent, nil, nil)
		}
	}
}

// Update is ran every frame, with `dt` being the time
// in seconds since the last frame
func (cb *CityBuildingSystem) Update(dt float32) {
	buildKey := engo.Input.Button("build")
	if buildKey.JustPressed() {
		fmt.Println("The gamer pressed Q")

		// Create a new city
		city := createCity(cb)

		// Add to the system
		for _, system := range cb.world.Systems() {
			switch sys := system.(type) {
			case *common.RenderSystem:
				sys.Add(&city.BasicEntity, &city.RenderComponent, &city.SpaceComponent)
			}
		}
	}
}

// Remove is called whenever an Entity is removed from the scene, and thus from this system
func (*CityBuildingSystem) Remove(ecs.BasicEntity) {}

// createCity creates a city
func createCity(cb *CityBuildingSystem) (city City) {
	texture, err := common.LoadedSprite(cityTextureLocation)
	if err != nil {
		panic(err)
	}

	city = City{BasicEntity: ecs.NewBasic()}
	city.RenderComponent = common.RenderComponent{
		Drawable: texture,
		Scale:    engo.Point{0.1, 0.1},
	}
	city.SpaceComponent = common.SpaceComponent{
		Position: engo.Point{cb.mouseTracker.MouseComponent.MouseX, cb.mouseTracker.MouseComponent.MouseY},
		Width:    30,
		Height:   64,
	}

	return city
}
