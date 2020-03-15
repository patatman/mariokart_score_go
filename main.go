package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	log "github.com/sirupsen/logrus"
)

type Game struct {
	gorm.Model
	Timestamp   time.Time `gorm:"not null;default:current_timestamp"`
	GamePlayers []*GamePlayer
}

type GamePlayer struct {
	gorm.Model
	Game     *Game
	GameID   uint `gorm:"not null"`
	Player   *Player
	PlayerID uint `gorm:"not null"`
	Score    uint `gorm:"not null;default:0"`
}

type Player struct {
	gorm.Model
	Name        string `gorm:"not null;unique"`
	DisplayName string `gorm:"not null;unique"`
	GamePlayers []*GamePlayer
}

type TotalScore struct {
	PlayerID uint
	Score    uint
}

func createPlayer(db *gorm.DB, name string, displayName string) (*Player, []error) {
	p := &Player{
		Name:        name,
		DisplayName: displayName,
	}
	db.Create(p)
	errs := db.GetErrors()
	if len(errs) > 0 {
		return nil, errs
	}
	return p, nil
}

func getPlayerByName(db *gorm.DB, name string) *Player {
	p := &Player{}
	db.Where(&Player{Name: name}).First(p)
	return p
}

func getPlayerByID(db *gorm.DB, playerID uint) *Player {
	p := &Player{}
	db.First(p, playerID)
	return p
}

func createGame(db *gorm.DB, ts time.Time) (*Game, []error) {
	g := &Game{Timestamp: ts}
	db.Create(g)
	errs := db.GetErrors()
	if len(errs) > 0 {
		return nil, errs
	}
	return g, nil
}

func addPlayerToGame(db *gorm.DB, g *Game, p *Player, score uint) (*GamePlayer, []error) {
	gp := &GamePlayer{Game: g, Player: p, Score: score}
	db.Create(gp)
	errs := db.GetErrors()
	if len(errs) > 0 {
		return nil, errs
	}
	return gp, nil
}

func logErrors(errs []error) {
	if errs != nil && len(errs) > 0 {
		log.Error("Error creating game:")
		for _, e := range errs {
			log.Error(e)
		}
	}
}

func getScoresByGame(db *gorm.DB, gameID uint) []*GamePlayer {
	var gps []*GamePlayer
	db.Where("game_id = ?", gameID).
		Preload("Player").
		Preload("Game").
		Order("score desc").
		Find(&gps)
	return gps
}

func getOverallScores(db *gorm.DB) []*TotalScore {
	var scores []*TotalScore
	db.Table("game_players").Select("player_id, sum(score) as score").Group("player_id").Order("score desc").Scan(&scores)
	return scores
}

// HTTP part

// Load index data
//func loadIndex() {
//	htmlMetadata := "hello"
//
//	return htmlMetadata
//}
//
//func indexHandler(w http.ResponseWriter, r *http.Request) {
//	p := loadIndex()
//	if err != nill {
//		http.Error(w, err.Error(), http.StatusInternalServerError)
//		return
//	}
//	renderTemplate(w, "index", p)
//}

func main() {
	// Enable debug logging...
	log.SetLevel(log.DebugLevel)

	db, err := gorm.Open("sqlite3", "./mariocart.sqlite")
	if err != nil {
		log.Fatalf("error creating database: %v", err)
	}
	log.Info("Database opened...")
	defer db.Close()

	// Enable logging for GORM
	db.LogMode(false)
	db.SetLogger(log.New())

	// Update table structure (if needed)
	db.AutoMigrate(&Player{})
	db.AutoMigrate(&Game{})
	db.AutoMigrate(&GamePlayer{})

	// Create a new player
	//jopie, errs := createPlayer(db, "jopie", "Sloper")
	//logErrors(errs)
	//log.Infof("Created player: %v", jopie)
	//
	// Query existing players
	//bas := getPlayerByName(db, "bastiaan")
	//pim := getPlayerByName(db, "pim")
	//serge := getPlayerByName(db, "serge")
	//log.Infof("P1: %v", bas)
	//log.Infof("P2: %v", pim)
	//log.Infof("P3: %v", serge)
	//
	// Create a new game
	//game, errs := createGame(db, time.Now())
	//logErrors(errs)
	//log.Infof("Game: %v", game)
	//
	// Add player-scores to a game
	//score1, kaka := addPlayerToGame(db, game, pim, 10000)
	//logErrors(kaka)
	//log.Infof("Score added: %v", score1)
	//
	//score2, booboo := addPlayerToGame(db, game, bas, 100)
	//logErrors(booboo)
	//log.Infof("Score added: %v", score2)
	//
	//score3, errs := addPlayerToGame(db, game, serge, 10)
	//logErrors(errs)
	//log.Infof("Score added: %v", score3)

	// Query player-scores from a game, using the game's ID
	//gps := getScoresByGame(db, 3)
	//for _, gp := range gps {
	//	log.Infof("%s == %s: %d", gp.Game.Timestamp, gp.Player.DisplayName, gp.Score)
	//}

	// Get the overall high scores
	//totals := getOverallScores(db)
	//for i, s := range totals {
	//	p := getPlayerByID(db, s.PlayerID)
	//	log.Infof("%d - %s (%s) - %d", i+1, p.DisplayName, p.Name, s.Score)
	//}

	// Server HTTP server
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		totals := getOverallScores(db)
		for i, s := range totals {
			p := getPlayerByID(db, s.PlayerID)
			fmt.Fprintf(w, "%d - %s (%s) - %d", i+1, p.DisplayName, p.Name, s.Score)
		}
	})
	http.ListenAndServe(":80", nil)
}
