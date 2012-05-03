package dk.frankbille.scoreboard.domain;

import java.io.Serializable;

public class GameTeam implements Serializable {
	private static final long serialVersionUID = 1L;

	private Long id;

	private Game game;

	private Team team;

	private int score;

	public Long getId() {
		return id;
	}

	public void setId(Long id) {
		this.id = id;
	}

	public Game getGame() {
		return game;
	}

	public void setGame(Game game) {
		this.game = game;
	}

	public Team getTeam() {
		return team;
	}

	public void setTeam(Team team) {
		this.team = team;
	}

	public int getScore() {
		return score;
	}

	public void setScore(int score) {
		this.score = score;
	}

	public boolean isWinner() {
		return getGame().didTeamWin(this);
	}

	public boolean hasPlayer(Player player) {
		boolean found = false;
		for (Player p : team.getPlayers()) {
			if (player.equals(p)) {
				found = true;
				break;
			}
		}
		return found;
	}
}
