package systems

import (
	"fmt"

	"engo.io/ecs"
	"engo.io/engo"
)

type MouseTracker struct {
	ecs.BasicEntity
	engo.MouseComponent
}

type City struct {
	ecs.BasicEntity
	engo.RenderComponent
	engo.SpaceComponent
}

type CityBuildingSystem struct {
	world *ecs.World

	mouseTracker MouseTracker
}

// Remove is called whenever an Entity is removed from the scene, and thus from this system
func (*CityBuildingSystem) Remove(ecs.BasicEntity) {}

// New is the initialisation of the System
func (cb *CityBuildingSystem) New(w *ecs.World) {
	cb.world = w
	fmt.Println("CityBuildingSystem was added to the Scene")

	cb.mouseTracker.BasicEntity = ecs.NewBasic()
	cb.mouseTracker.MouseComponent = engo.MouseComponent{Track: true}

	for _, system := range w.Systems() {
		switch sys := system.(type) {
		case *engo.MouseSystem:
			sys.Add(&cb.mouseTracker.BasicEntity, &cb.mouseTracker.MouseComponent, nil, nil)
		}
	}
}

// Update is ran every frame, with `dt` being the time
// in seconds since the last frame
func (cb *CityBuildingSystem) Update(dt float32) {
	if engo.Keys.Get(engo.F1).JustPressed() {
		fmt.Println("The gamer pressed F1")

		city := City{BasicEntity: ecs.NewBasic()}

		city.SpaceComponent = engo.SpaceComponent{
			Position: engo.Point{cb.mouseTracker.MouseComponent.MouseX, cb.mouseTracker.MouseComponent.MouseY},
			Width:    30,
			Height:   64,
		}

		texture := engo.Files.Image("city.png")
		city.RenderComponent = engo.NewRenderComponent(
			texture,
			engo.Point{0.1, 0.1},
			"city texture",
		)

		for _, system := range cb.world.Systems() {
			switch sys := system.(type) {
			case *engo.RenderSystem:
				sys.Add(&city.BasicEntity, &city.RenderComponent, &city.SpaceComponent)
			}
		}
	}
}
