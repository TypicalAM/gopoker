import React from 'react';

enum round {
	PreFlop = 'preflop',
	Flop = 'flop',
	Turn = 'turn',
	River = 'river',
}

interface Player {
	Name: string;
	Active: boolean;
	Action: string;

	Assets: number;
	Bet: number;
	HoleCards: string[];
}

interface GameState {
	ActiveBet: number;
	Pot: number;
	Round: round;
	CurrentPlayer: number;

	CommunityCards: null | string[];
	Players: Player[];

	BestHand: null | string[];
	BestRank: string;
	GameWinner: string;
}

const DefaultGameState: GameState = {
	ActiveBet: 0,
	Pot: 0,
	Round: round.PreFlop,
	CurrentPlayer: 0,
	CommunityCards: ["Ah", "Ks", "Qs", "Js", "Ts"],
	Players: [
		{
			Name: 'Player 1',
			Active: true,
			Action: 'none',
			Assets: 1000,
			Bet: 0,
			HoleCards: ["Ah", "Ks"],
		},
		{
			Name: 'Player 2',
			Active: true,
			Action: 'call',
			Assets: 1000,
			Bet: 2,
			HoleCards: [],
		},
		{
			Name: 'Player 3',
			Active: true,
			Action: 'call',
			Assets: 1000,
			Bet: 1,
			HoleCards: [],
		}
	],
	BestHand: null,
	BestRank: '',
	GameWinner: '',
};

export { round, DefaultGameState };
export type { Player, GameState };
