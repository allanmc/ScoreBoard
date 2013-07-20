package scoreboard

import (
	"appengine"
	"appengine/datastore"
	"time"
)

func (r ScoreBoardService) HandleGetLeagueGames() []Game {
	c := appengine.NewContext(r.Request())
	return LoadLeagueGames(c, r.Vars()["leagueId"])
}

type EditGame struct {
	Id       int64        `json:"id"`
	GameDate time.Time    `json:"gameDate"`
	League   string       `json:"league"`
	Team1    EditGameTeam `json:"team1"`
	Team2    EditGameTeam `json:"team2"`
}

type EditGameTeam struct {
	Score   int32    `json:"score"`
	Players []string `json:"players"`
}

func (r ScoreBoardService) HandleSaveGame(editGame EditGame) Game {
	c := appengine.NewContext(r.Request())

	// Zero out the time
	year, month, day := editGame.GameDate.Date()
	editGame.GameDate = time.Date(year, month, day, 0, 0, 0, 0, time.UTC)

	var key *datastore.Key
	
	err := datastore.RunInTransaction(c, func(c appengine.Context) error {
		leagueKey := datastore.NewKey(c, ENTITY_LEAGUE, editGame.League, 0, nil)

		var game Game
		
		if editGame.Id == 0 {
			key = datastore.NewIncompleteKey(c, ENTITY_GAME, leagueKey)

			game = Game{
				Team1: GameTeam{},
				Team2: GameTeam{},
			}
		} else {
			key = datastore.NewKey(c, ENTITY_GAME, "", editGame.Id, leagueKey)

			err := datastore.Get(c, key, &game)
			if err != nil {
				return err
			}
		}

		game.GameDate = editGame.GameDate
		game.League = leagueKey
		copyGameTeam(&game.Team1, editGame.Team1)
		copyGameTeam(&game.Team2, editGame.Team2)
		game.ChangeDate = time.Now().UnixNano()

		var err error
		key, err = datastore.Put(c, key, &game)
		if err != nil {
			return err
		}

		return nil
	}, nil)
	if err != nil {
		c.Errorf("%v", err)
	}

	err = RecalculateLeagueRatings(c, editGame.League)
	if err != nil {
		c.Errorf("%v", err)
	}
	
	var game Game
	
	err = datastore.Get(c, key, &game)
	if err != nil {
		c.Errorf("%v", err)
	}
	
	game.Id = key.IntID()
	
	return game
}

func copyGameTeam(gameTeam *GameTeam, editGameTeam EditGameTeam) {
	gameTeam.Score = editGameTeam.Score
	gameTeam.Players = make([]TeamPlayer, len(editGameTeam.Players))
	for i, player := range editGameTeam.Players {
		gameTeam.Players[i] = TeamPlayer{
			Player: player,
		}
	}
}

func RecalculateLeagueRatings(c appengine.Context, leagueId string) error {
	leagueKey := datastore.NewKey(c, ENTITY_LEAGUE, leagueId, 0, nil)
	gamesQuery := datastore.NewQuery(ENTITY_GAME).Ancestor(leagueKey).Order("GameDate").Order("ChangeDate")
	var games []Game
	gameKeys, err := gamesQuery.GetAll(c, &games)
	if err != nil {
		return err
	}

	playerRatings := map[string]float64{}

	for i := 0; i < len(games); i++ {
		game := games[i]
		team1Score := game.Team1.Score
		team2Score := game.Team2.Score

		team1Rating := getTeamRating(game.Team1.Players, playerRatings)
		team2Rating := getTeamRating(game.Team2.Players, playerRatings)

		var resultRating float64
		if team1Score > team2Score {
			resultRating = CalculateRating(team1Rating, team2Rating, team1Score, team2Score)
			applyResultRating(c, &game.Team1, playerRatings, resultRating, true)
			applyResultRating(c, &game.Team2, playerRatings, resultRating, false)
		} else {
			resultRating = CalculateRating(team2Rating, team1Rating, team2Score, team1Score)
			applyResultRating(c, &game.Team2, playerRatings, resultRating, true)
			applyResultRating(c, &game.Team1, playerRatings, resultRating, false)
		}
	}

	for i := 0; i < len(games); i += 500 {
		start := i
		end := i + 500
		if end > len(games) {
			end = len(games)
		}
		_, err := datastore.PutMulti(c, gameKeys[start:end], games[start:end])

		if err != nil {
			return err
		}
	}

	return nil
}

func getTeamRating(teamPlayers []TeamPlayer, playerRatings map[string]float64) float64 {
	var teamRating float64

	for i := 0; i < len(teamPlayers); i++ {
		teamPlayer := teamPlayers[i]
		playerRating, found := playerRatings[teamPlayer.Player]
		if found == false {
			playerRating = 1000
			playerRatings[teamPlayer.Player] = playerRating
		}

		teamRating += playerRating
	}

	return teamRating / float64(len(teamPlayers))
}

func applyResultRating(c appengine.Context, gameTeam *GameTeam, playerRatings map[string]float64, resultRating float64, winner bool) {
	teamPlayers := gameTeam.Players

	perPlayerRating := resultRating / float64(len(teamPlayers))

	for i := 0; i < len(teamPlayers); i++ {
		teamPlayer := &teamPlayers[i]
		startRating := playerRatings[teamPlayer.Player]
		var endRating float64
		if winner {
			endRating = startRating + perPlayerRating
		} else {
			endRating = startRating - perPlayerRating
		}
		playerRatings[teamPlayer.Player] = endRating

		teamPlayer.StartRating = startRating
		teamPlayer.EndRating = endRating
	}
}
