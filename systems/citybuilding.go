package systems

import (
	"fmt"
	"image/color"
	"math/rand"

	"engo.io/ecs"
	"engo.io/engo"
	"engo.io/engo/common"
)

const (
	cityZIndex = 100
)

type MouseTracker struct {
	ecs.BasicEntity
	common.MouseComponent
}

type City struct {
	ecs.BasicEntity
	common.RenderComponent
	common.SpaceComponent
	common.MouseComponent
	CityComponent
}

type CityComponent struct {
	Name       string
	Population int

	Roads []*Road

	isHovered bool
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

	engo.Input.RegisterButton("build", engo.F1, engo.B, engo.ArrowDown)

	cb.mouseTracker.BasicEntity = ecs.NewBasic()
	cb.mouseTracker.MouseComponent = common.MouseComponent{Track: true}

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
	if engo.Input.Button("build").JustPressed() {
		city := City{BasicEntity: ecs.NewBasic()}

		city.SpaceComponent = common.SpaceComponent{
			Position: engo.Point{cb.mouseTracker.MouseComponent.MouseX, cb.mouseTracker.MouseComponent.MouseY},
			Width:    30,
			Height:   30,
		}

		city.RenderComponent = common.RenderComponent{
			Drawable: common.Rectangle{},
			Color:    color.Black,
		}
		city.RenderComponent.SetZIndex(cityZIndex)
		city.RenderComponent.SetShader(common.LegacyShader)

		city.CityComponent = CityComponent{
			Name:       fmt.Sprintf("City %d", city.BasicEntity.ID()),
			Population: rand.Intn(500),
		}

		for _, system := range cb.world.Systems() {
			switch sys := system.(type) {
			case *common.RenderSystem:
				sys.Add(&city.BasicEntity, &city.RenderComponent, &city.SpaceComponent)
			case *common.MouseSystem:
				sys.Add(&city.BasicEntity, &city.MouseComponent, &city.SpaceComponent, &city.RenderComponent)
			case *RoadBuildingSystem:
				sys.AddCity(&city.BasicEntity, &city.CityComponent, &city.SpaceComponent, &city.RenderComponent, &city.MouseComponent)
			case *HUDSystem:
				sys.AddCity(&city.BasicEntity, &city.CityComponent, &city.MouseComponent)
			case *CommuterSystem:
				sys.AddCity(&city.BasicEntity, &city.CityComponent)
			}
		}
	}
}
