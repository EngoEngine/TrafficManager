package systems

import (
	"fmt"
	"image/color"
	"sync/atomic"

	"engo.io/ecs"
	"engo.io/engo"
	"engo.io/engo/common"
	"github.com/EngoEngine/TrafficDefense/systems/ui"
)

const (
	defaultMoney int64 = 1000000

	moneySize                = clockSize
	moneyPadding             = clockPadding
	moneyZIndex              = clockZIndex
	moneyMarginRight float32 = 30
)

type MoneyComponent struct {
	amount int64
}

func (m *MoneyComponent) Add(a int64) {
	atomic.AddInt64(&m.amount, a)
}

func (m *MoneyComponent) Amount() int64 {
	return m.amount
}

type moneyEntityClock struct {
	*ecs.BasicEntity
	*common.SpaceComponent
}

type money struct {
	MoneyComponent
	ui.Label
}

type MoneySystem struct {
	money money

	clock moneyEntityClock
}

func (m *MoneySystem) New(w *ecs.World) {
	// Default values
	m.money.amount = defaultMoney

	// Load the preloaded font
	fnt := &common.Font{
		URL:  robotoFontLocation,
		FG:   color.Black,
		Size: moneySize,
	}
	if err := fnt.CreatePreloaded(); err != nil {
		panic(err)
	}

	// Create the visual money
	m.money.BasicEntity = ecs.NewBasic()

	m.money.Font = fnt
	m.money.RenderComponent.Color = color.Black
	m.money.SetText("$ ")
	m.money.SpaceComponent = common.SpaceComponent{
		Position: engo.Point{
			X: engo.CanvasWidth() - 2*m.money.Drawable.Width() - moneyPadding - moneyMarginRight,
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

		m.money.Add(int64(cmtr.Commuter.Vehicle.Reward))
	})
}

func (m *MoneySystem) Remove(basic ecs.BasicEntity) {
	if basic.ID() == m.clock.ID() {
		m.clock = moneyEntityClock{}
	}
}

func (m *MoneySystem) SetClock(basic *ecs.BasicEntity, space *common.SpaceComponent) {
	m.clock = moneyEntityClock{basic, space}
}

func (m *MoneySystem) Update(dt float32) {
	if m.money.SetText(fmt.Sprintf("$ %v", m.money.amount)) {
		m.money.Width = m.money.Drawable.Width()
	}
	m.money.Position.X = engo.CanvasWidth() - m.money.Width - m.clock.Width - moneyMarginRight
}
