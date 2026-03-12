// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"image/color"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

const (
	// Board dimensions (in tiles)
	boardCols = 28
	boardRows = 31

	// Tile size in pixels
	tileSize = 16

	// Window dimensions
	screenWidth  = boardCols * tileSize // 448
	screenHeight = boardRows * tileSize // 496

	// Game title
	title = "Pac-Man"
)

// Direction represents movement direction.
type Direction int

const (
	DirNone Direction = iota
	DirUp
	DirDown
	DirLeft
	DirRight
)

// Game implements the ebiten.Game interface.
type Game struct {
	// Pac-Man position (in pixels)
	playerX float64
	playerY float64

	// Current movement direction
	playerDir Direction

	// Score
	score int
}

// NewGame creates and returns a new Game instance.
func NewGame() *Game {
	return &Game{
		playerX:   float64(14 * tileSize),
		playerY:   float64(23 * tileSize),
		playerDir: DirNone,
		score:     0,
	}
}

// Update handles game logic each tick.
func (g *Game) Update() error {
	switch {
	case ebiten.IsKeyPressed(ebiten.KeyArrowUp):
		g.playerDir = DirUp
	case ebiten.IsKeyPressed(ebiten.KeyArrowDown):
		g.playerDir = DirDown
	case ebiten.IsKeyPressed(ebiten.KeyArrowLeft):
		g.playerDir = DirLeft
	case ebiten.IsKeyPressed(ebiten.KeyArrowRight):
		g.playerDir = DirRight
	}

	speed := 2.0
	switch g.playerDir {
	case DirUp:
		g.playerY -= speed
	case DirDown:
		g.playerY += speed
	case DirLeft:
		g.playerX -= speed
	case DirRight:
		g.playerX += speed
	}

	if g.playerX < 0 {
		g.playerX = 0
	}
	if g.playerY < 0 {
		g.playerY = 0
	}
	if g.playerX > float64(screenWidth-tileSize) {
		g.playerX = float64(screenWidth - tileSize)
	}
	if g.playerY > float64(screenHeight-tileSize) {
		g.playerY = float64(screenHeight - tileSize)
	}

	return nil
}

// Draw renders the game screen.
func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.Black)

	pacmanColor := color.RGBA{R: 255, G: 255, B: 0, A: 255}
	for y := 0; y < tileSize; y++ {
		for x := 0; x < tileSize; x++ {
			px := int(g.playerX) + x
			py := int(g.playerY) + y
			if px >= 0 && px < screenWidth &&
				py >= 0 && py < screenHeight {
				screen.Set(px, py, pacmanColor)
			}
		}
	}

	ebitenutil.DebugPrint(
		screen,
		fmt.Sprintf(
			"Score: %d\nTPS: %0.2f\nArrow keys to move",
			g.score,
			ebiten.ActualTPS(),
		),
	)
}

// Layout returns the game's logical screen dimensions.
func (g *Game) Layout(
	outsideWidth, outsideHeight int,
) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	ebiten.SetWindowSize(screenWidth*2, screenHeight*2)
	ebiten.SetWindowTitle(title)

	game := NewGame()
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
