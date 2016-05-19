package systems

import (
	"fmt"
	"image/color"
	"sync/atomic"

	"engo.io/ecs"
	"engo.io/engo"
	"engo.io/engo/common"
)

const (
	defaultMoney int64 = 1000000

	moneySize    = clockSize
	moneyPadding = clockPadding
	moneyZIndex  = clockZIndex
)

type MoneyComponent struct {
	amount int64
}

// Add adds more money to the
func (m *MoneyComponent) Add(a int64) {
	atomic.AddInt64(&m.amount, a)
}

func (m *MoneyComponent) Amount() int64 {
	return m.amount
}

type money struct {
	ecs.BasicEntity
	MoneyComponent
	common.RenderComponent
	common.SpaceComponent
}

type MoneySystem struct {
	money      money
	moneyCache int64

	robotoFont common.Font
}

func (m *MoneySystem) New(w *ecs.World) {
	// Default values
	m.money.amount = defaultMoney

	// Load the preloaded font
	m.robotoFont = common.Font{
		URL:  robotoFontLocation,
		FG:   color.Black,
		Size: moneySize,
	}
	if err := m.robotoFont.CreatePreloaded(); err != nil {
		panic(err)
	}

	// Create the visual money
	m.money.BasicEntity = ecs.NewBasic()

	m.money.RenderComponent = common.RenderComponent{
		Drawable: m.robotoFont.Render(fmt.Sprintf("$ %v", m.money.amount)),
		Color:    color.Black,
	}
	m.money.SpaceComponent = common.SpaceComponent{
		Position: engo.Point{
			X: engo.CanvasWidth() - 2*m.money.Drawable.Width() - moneyPadding,
			Y: moneyPadding,
		},
		Width:  m.money.Drawable.Width(),
		Height: m.money.Drawable.Height() + 2*moneyPadding,
	}
	m.money.SetZIndex(moneyZIndex)
	m.money.SetShader(common.HUDShader)

	for _, system := range w.Systems() {
		switch sys := system.(type) {
		case *common.RenderSystem:
			sys.Add(&m.money.BasicEntity, &m.money.RenderComponent, &m.money.SpaceComponent)
		case *RoadBuildingSystem:
			sys.SetMoney(&m.money.BasicEntity, &m.money.MoneyComponent)
		}
	}

	engo.Mailbox.Listen("CommuterArrivedMessage", func(msg engo.Message) {
		cmtr, ok := msg.(CommuterArrivedMessage)
		if !ok {
			return
		}

		m.money.Add(int64(cmtr.Commuter.DistanceTravelled * cashPerUnit))
	})
}

func (*MoneySystem) Remove(ecs.BasicEntity) {}

func (m *MoneySystem) Update(dt float32) {
	if m.money.amount != m.moneyCache {
		m.money.Drawable.Close()
		m.money.Drawable = m.robotoFont.Render(fmt.Sprintf("$ %v", m.money.amount))
		m.money.Width = m.money.Drawable.Width()
		m.moneyCache = m.money.amount
	}
	m.money.Position.X = engo.CanvasWidth() - m.money.Width // todo: minus c.clock.Width
}
