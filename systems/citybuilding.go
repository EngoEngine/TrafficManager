package systems

import (
	"fmt"

	"engo.io/ecs"
	"engo.io/engo"
)

type CityBuildingSystem struct {
	world *ecs.World

	mouseTracker *ecs.Entity
}

// Type returns a unique string identifier, usually
// something like "RenderSystem", "CollisionSystem"...
func (*CityBuildingSystem) Type() string {
	return "CityBuildingSystem"
}

// Priority is used to create the order in which Systems
// (in the World) are processed
func (*CityBuildingSystem) Priority() int {
	return 0
}

// AddEntity is called whenever an Entity is added to the
// World, which "requires" this System
func (*CityBuildingSystem) AddEntity(*ecs.Entity) {}

// RemoveEntity is called whenever an Entity is removed from
// the World, which "requires" this System
func (*CityBuildingSystem) RemoveEntity(*ecs.Entity) {}

// New is the initialisation of the System
func (cb *CityBuildingSystem) New(w *ecs.World) {
	cb.world = w
	fmt.Println("CityBuildingSystem was added to the Scene")

	cb.mouseTracker = ecs.NewEntity([]string{"MouseSystem"})
	cb.mouseTracker.AddComponent(&engo.MouseComponent{Track: true})
	w.AddEntity(cb.mouseTracker)
}

// Update is ran every frame, with `dt` being the time
// in seconds since the last frame
func (cb *CityBuildingSystem) Update(dt float32) {
	if engo.Keys.Get(engo.F1).JustPressed() {
		fmt.Println("The gamer pressed F1")
		entity := ecs.NewEntity([]string{"RenderSystem"})

		var (
			mouse *engo.MouseComponent
			ok bool
		)

		if mouse, ok = cb.mouseTracker.ComponentFast(mouse).(*engo.MouseComponent); !ok {
			return
		}

		entity.AddComponent(&engo.SpaceComponent{
			Position: engo.Point{mouse.MouseX, mouse.MouseY},
			Width: 30,
			Height: 64,
		})

		texture := engo.Files.Image("city.png")
		entity.AddComponent(engo.NewRenderComponent(
			texture,
			engo.Point{0.1, 0.1},
			"city texture",
		))

		cb.world.AddEntity(entity)
	}
}
