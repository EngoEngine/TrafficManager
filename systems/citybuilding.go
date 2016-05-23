package systems

import (
	"image/color"

	"engo.io/ecs"
	"engo.io/engo"
	"engo.io/engo/common"
)

const (
	cityZIndex = 100

	CityWidth  float32 = 30
	CityHeight float32 = 30
)

type MouseTracker struct {
	ecs.BasicEntity
	common.MouseComponent
}

type CityCategory uint8

const (
	CategoryRed CityCategory = iota
	CategoryGreen
	CategoryBlue
)

func (c CityCategory) String() string {
	switch c {
	case CategoryRed:
		return "Red"
	case CategoryGreen:
		return "Green"
	case CategoryBlue:
		return "Blue"
	default:
		panic("CityCategory not found for String() method")
	}
}

func (c CityCategory) Color() color.Color {
	switch c {
	case CategoryRed:
		return color.RGBA{255, 0, 0, 255}
	case CategoryGreen:
		return color.RGBA{0, 255, 0, 255}
	case CategoryBlue:
		return color.RGBA{0, 0, 255, 255}
	default:
		panic("CityCategory not found for String() method")
	}
}

type City struct {
	ecs.BasicEntity
	common.RenderComponent
	common.SpaceComponent
	common.MouseComponent
	CityComponent
}

type CityComponent struct {
	Category CityCategory

	Queue      []WaveComponent
	RedCount   int
	GreenCount int
	BlueCount  int

	Roads     []*Road
	isHovered bool
}

func (c *CityComponent) Enqueue(wave []WaveComponent) {
	c.Queue = wave
	for _, elem := range wave {
		switch elem.To {
		case CategoryRed:
			c.RedCount++
		case CategoryGreen:
			c.GreenCount++
		case CategoryBlue:
			c.BlueCount++
		}
	}
}

func BuildCity(x, y float32, cat CityCategory, w *ecs.World) *City {
	city := &City{
		BasicEntity:   ecs.NewBasic(),
		CityComponent: CityComponent{Category: cat},
		SpaceComponent: common.SpaceComponent{
			Position: engo.Point{x, y},
			Width:    CityWidth,
			Height:   CityHeight,
		},
		RenderComponent: common.RenderComponent{
			Drawable: common.Rectangle{},
			Color:    cat.Color(),
		},
	}

	city.RenderComponent.SetZIndex(cityZIndex)
	city.RenderComponent.SetShader(common.LegacyShader)

	for _, system := range w.Systems() {
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
		case *WaveSystem:
			sys.AddCity(&city.BasicEntity, &city.CityComponent)
		}
	}

	return city
}
