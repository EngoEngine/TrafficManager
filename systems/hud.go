package systems

import (
	"fmt"
	"image/color"

	"engo.io/ecs"
	"engo.io/engo"
	"engo.io/engo/common"
	"github.com/EngoEngine/TrafficManager/systems/ui"
)

var (
	hudZ                = float32(1000)
	hudHeight           = float32(200)
	hudCityTitlePadding = float32(4)
)

type hudEntityCity struct {
	*ecs.BasicEntity
	*CityComponent
	*common.MouseComponent
}

type HUDSystem struct {
	world *ecs.World

	cities []hudEntityCity

	hudFrame     ui.Graphic
	hudCityTitle ui.Label
}

func (h *HUDSystem) Remove(basic ecs.BasicEntity) {
	delete := -1
	for index, e := range h.cities {
		if e.BasicEntity.ID() == basic.ID() {
			delete = index
			break
		}
	}
	if delete >= 0 {
		h.cities = append(h.cities[:delete], h.cities[delete+1:]...)
	}
}

func (h *HUDSystem) New(w *ecs.World) {
	h.world = w

	h.hudFrame = ui.Graphic{BasicEntity: ecs.NewBasic()}
	h.hudFrame.SpaceComponent = common.SpaceComponent{
		Position: engo.Point{0, engo.CanvasHeight() - hudHeight},
		Width:    engo.CanvasWidth(),
		Height:   hudHeight,
	}

	h.hudFrame.RenderComponent = common.RenderComponent{
		Drawable: common.Rectangle{BorderWidth: 1, BorderColor: color.RGBA{20, 20, 20, 255}},
		Color:    color.RGBA{200, 200, 200, 255},
	}
	h.hudFrame.RenderComponent.SetZIndex(hudZ)
	h.hudFrame.RenderComponent.SetShader(common.LegacyHUDShader)

	for _, system := range h.world.Systems() {
		switch sys := system.(type) {
		case *common.RenderSystem:
			sys.Add(&h.hudFrame.BasicEntity, &h.hudFrame.RenderComponent, &h.hudFrame.SpaceComponent)
		}
	}

	fnt := &common.Font{
		URL:  "fonts/Roboto-Regular.ttf",
		FG:   color.Black,
		Size: 24,
	}
	err := fnt.CreatePreloaded()
	if err != nil {
		panic(err)
	}

	h.hudCityTitle = ui.Label{BasicEntity: ecs.NewBasic()}
	h.hudCityTitle.SpaceComponent = common.SpaceComponent{
		Position: engo.Point{hudCityTitlePadding, engo.CanvasHeight() - hudHeight},
		Width:    engo.CanvasWidth(),
		Height:   hudHeight,
	}

	h.hudCityTitle.Font = fnt
	h.hudCityTitle.RenderComponent.SetZIndex(hudZ + 1)
	h.hudCityTitle.RenderComponent.SetShader(common.HUDShader)

	for _, system := range h.world.Systems() {
		switch sys := system.(type) {
		case *common.RenderSystem:
			sys.Add(&h.hudCityTitle.BasicEntity, &h.hudCityTitle.RenderComponent, &h.hudCityTitle.SpaceComponent)
		}
	}
}

func (h *HUDSystem) AddCity(basic *ecs.BasicEntity, city *CityComponent, mouse *common.MouseComponent) {
	h.cities = append(h.cities, hudEntityCity{basic, city, mouse})
}

func (h *HUDSystem) Update(dt float32) {
	// Possibly update the location
	h.hudFrame.SpaceComponent.Position.Y = engo.CanvasHeight() - hudHeight
	h.hudFrame.Width = engo.CanvasWidth()
	h.hudCityTitle.SpaceComponent.Position.Y = engo.CanvasHeight() - hudHeight + hudCityTitlePadding
	h.hudCityTitle.Width = engo.CanvasWidth()/2 - 2*hudCityTitlePadding

	// Update the text shown
	var cityHovered bool
	for _, city := range h.cities {
		if city.MouseComponent.Hovered {
			cityHovered = true
			h.hudCityTitle.SetText(fmt.Sprintf("City %q", city.CityComponent.Category))
			break
		}
	}

	h.hudCityTitle.Hidden = !cityHovered
}
