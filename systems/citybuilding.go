package systems

import (
	"github.com/EngoEngine/ecs"
	"github.com/EngoEngine/engo"
	"github.com/EngoEngine/engo/common"

	"math/rand"
	"time"
)

// Spritesheet contains the sprites for the city buildings, cars, and roads
var Spritesheet *common.Spritesheet

var cities = [...][12]int{
	{99, 100, 101,
		454, 269, 455,
		415, 195, 416,
		452, 306, 453,
	},
	{99, 100, 101,
		268, 269, 270,
		268, 269, 270,
		305, 306, 307,
	},
	{75, 76, 77,
		446, 261, 447,
		446, 261, 447,
		444, 298, 445,
	},
	{75, 76, 77,
		407, 187, 408,
		407, 187, 408,
		444, 298, 445,
	},
	{75, 76, 77,
		186, 150, 188,
		186, 150, 188,
		297, 191, 299,
	},
	{83, 84, 85,
		413, 228, 414,
		411, 191, 412,
		448, 302, 449,
	},
	{83, 84, 85,
		227, 228, 229,
		190, 191, 192,
		301, 302, 303,
	},
	{91, 92, 93,
		241, 242, 243,
		278, 279, 280,
		945, 946, 947,
	},
	{91, 92, 93,
		241, 242, 243,
		278, 279, 280,
		945, 803, 947,
	},
	{91, 92, 93,
		238, 239, 240,
		238, 239, 240,
		312, 313, 314,
	},
}

// City is a city tile
type City struct {
	ecs.BasicEntity
	common.RenderComponent
	common.SpaceComponent
}

// CityBuildingSystem builds random cities as the game progresses
type CityBuildingSystem struct {
	world *ecs.World

	usedTiles []int

	elapsed, buildTime float32
	built              int
}

// New is the initialisation of the System
func (cb *CityBuildingSystem) New(w *ecs.World) {
	cb.world = w

	Spritesheet = common.NewSpritesheetWithBorderFromFile("textures/citySheet.png", 16, 16, 1, 1)

	rand.Seed(time.Now().UnixNano())

	cb.updateBuildTime()
}

// Update is ran every frame, with `dt` being the time
// in seconds since the last frame
func (cb *CityBuildingSystem) Update(dt float32) {
	cb.elapsed += dt
	if cb.elapsed >= cb.buildTime {
		cb.generateCity()
		cb.elapsed = 0
		cb.updateBuildTime()
		cb.built++
	}
}

// Remove is called whenever an Entity is removed from the scene, and thus from this system
func (*CityBuildingSystem) Remove(ecs.BasicEntity) {}

// generateCity randomly generates a city in a random location on the map
func (cb *CityBuildingSystem) generateCity() {
	x := rand.Intn(18)
	y := rand.Intn(18)
	t := x + y*18

	for cb.isTileUsed(t) {
		if len(cb.usedTiles) > 300 {
			break //to avoid infinite loop
		}
		x = rand.Intn(18)
		y = rand.Intn(18)
		t = x + y*18
	}
	cb.usedTiles = append(cb.usedTiles, t)

	city := rand.Intn(len(cities))
	cityTiles := make([]*City, 0)
	for i := 0; i < 3; i++ {
		for j := 0; j < 4; j++ {
			tile := &City{BasicEntity: ecs.NewBasic()}
			tile.SpaceComponent.Position = engo.Point{
				X: float32(((x+1)*64)+8) + float32(i*16),
				Y: float32(((y + 1) * 64)) + float32(j*16),
			}
			tile.RenderComponent.Drawable = Spritesheet.Cell(cities[city][i+3*j])
			tile.RenderComponent.SetZIndex(1)
			cityTiles = append(cityTiles, tile)
		}
	}

	for _, system := range cb.world.Systems() {
		switch sys := system.(type) {
		case *common.RenderSystem:
			for _, v := range cityTiles {
				sys.Add(&v.BasicEntity, &v.RenderComponent, &v.SpaceComponent)
			}
		}
	}

	engo.Mailbox.Dispatch(HUDTextMessage{
		BasicEntity: ecs.NewBasic(),
		SpaceComponent: common.SpaceComponent{
			Position: engo.Point{X: float32((x + 1) * 64), Y: float32((y + 1) * 64)},
			Width:    64,
			Height:   64,
		},
		MouseComponent: common.MouseComponent{},
		Line1:          "Town",
		Line2:          "Just built!",
		Line3:          "A town generates",
		Line4:          "$100 per day.",
	})

	engo.Mailbox.Dispatch(CityUpdateMessage{
		New: CityTypeNew,
	})
}

func (cb *CityBuildingSystem) isTileUsed(tile int) bool {
	for _, t := range cb.usedTiles {
		if tile == t {
			return true
		}
	}
	return false
}

func (cb *CityBuildingSystem) updateBuildTime() {
	switch {
	case cb.built < 2:
		cb.buildTime = 5*rand.Float32() + 10 // 10 to 15 seconds
	case cb.built < 8:
		cb.buildTime = 30*rand.Float32() + 60 // 60 to 90 seconds
	case cb.built < 18:
		cb.buildTime = 60*rand.Float32() + 30 // 30 to 90 seconds
	case cb.built < 28:
		cb.buildTime = 35*rand.Float32() + 30 // 30 to 65 seconds
	case cb.built < 33:
		cb.buildTime = 30*rand.Float32() + 30 // 30 to 60 seconds
	default:
		cb.buildTime = 20*rand.Float32() + 20 // 20 to 40 seconds
	}
}
