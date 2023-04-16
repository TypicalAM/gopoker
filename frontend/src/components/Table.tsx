import React, { useEffect } from 'react';
import { GameState } from './GameState';
import PlayerCard from './Player';
import PlayingCard from './Card';
import { GameMessage, MsgType } from './GameMessage';

interface TableProps {
	state: GameState;
	conn: WebSocket | null;
}

function Table(props: TableProps) {
	const [myIndex, setMyIndex] = React.useState(-1);
	const [winnerIndex, setWinnerIndex] = React.useState(-1);
	const [communityCards, setCommunityCards] = React.useState<string[]>([""]);

	useEffect(() => {
		for (let i = 0; i < props.state.Players.length; i++) {
			if (props.state.Players[i].HoleCards.length > 0) {
				console.log("My index is " + i);
				setMyIndex(i);
			}
		}

		// Pad the community cards so there is 5 strings
		let cards = props.state.CommunityCards;
		if (!cards) {
			cards = [];
		}

		while (cards.length < 5) {
			cards.push("");
		}

		if (props.state.GameOver) {
			for (let i = 0; i < props.state.Players.length; i++) {
				if (props.state.Players[i].Name === props.state.GameWinner) {
					setWinnerIndex(i);
					break;
				}
			}
		}

		setCommunityCards(cards);
	}, [props.state])

	return (
		<div className="w-full p-4 h-max">
			<div className="flex flex-col bg-gradient-to-br from-gray-800 to-gray-700 h-full w-full rounded-xl space-y-10 pb-5">
				<div className="top-0 left-0 pl-4 pt-2">
					<h1 className="text-xl font-bold text-gray-900 dark:text-gray-100">
						Pot <span className="ml-2 text-red-500">{props.state.Pot}</span>
					</h1>
				</div>

				<div className="flex flex-col items-center space-y-4">
					<h1 className="text-2xl font-bold text-gray-900 dark:text-gray-100">
						Current Bet <span className="ml-2 text-emerald-500">{props.state.ActiveBet}</span>
					</h1>
					<div className="flex justify-center space-x-4 bg-gray-800 rounded-xl p-3">
						{communityCards.map((card, _) => {
							if (!card) {
								return <PlayingCard {...{ Value: null, IsCommunity: true }} />
							} else {
								return <PlayingCard {...{ Value: card, IsCommunity: true }} />
							}
						})
						}
					</div>
				</div>

				<div className="flex flex-row justify-center items-center">

					<PlayerCard {...{
						Value: props.state.Players[0],
						Active: 0 === props.state.CurrentPlayer,
						IsMe: myIndex === 0,
						HasWon: 0 === winnerIndex,
						GameOver: props.state.GameOver
					}} />

					<PlayerCard {...{
						Value: props.state.Players[1],
						Active: 1 === props.state.CurrentPlayer,
						IsMe: myIndex === 1,
						HasWon: 1 === winnerIndex,
						GameOver: props.state.GameOver
					}} />

					<PlayerCard {...{
						Value: props.state.Players[2],
						Active: 2 === props.state.CurrentPlayer,
						IsMe: myIndex === 2,
						HasWon: 2 === winnerIndex,
						GameOver: props.state.GameOver
					}} />

				</div>

				{
					myIndex !== -1 && props.state.CurrentPlayer === myIndex ? (
						<div className="flex justify-end h-14 mb-4">
							<button className="mr-4 py-2 px-10 rounded text-white font-bold bg-gradient-to-br from-red-400 to-red-500 hover:bg-gradient-to-br hover:from-red-500 hover:to-red-500" onClick={() => {
								if (props.conn) {
									let mess: GameMessage = { type: MsgType.Action, data: "fold" }
									props.conn.send(JSON.stringify(mess))
								}
							}}> Fold </button>
							<button className="mr-4 py-2 px-10 rounded text-white font-bold bg-gray-800 hover:bg-gray-900" onClick={() => {
								if (props.conn) {
									let mess: GameMessage = { type: MsgType.Action, data: "check" }
									props.conn.send(JSON.stringify(mess))
								}
							}}>Check</button>
							<button className="mr-4 py-2 px-10 rounded text-white font-bold bg-gray-800 hover:bg-gray-900" onClick={() => {
								if (props.conn) {
									let mess: GameMessage = { type: MsgType.Action, data: "call" }
									props.conn.send(JSON.stringify(mess))
								}
							}}>Call</button>
							<button className="mr-4 py-2 px-10 rounded text-white font-bold bg-gradient-to-br from-emerald-400 to-emerald-500 hover:bg-gradient-to-br hover:from-emerald-500 hover:to-emerald-500" onClick={() => {
								if (props.conn) {
									let mess: GameMessage = { type: MsgType.Action, data: "raise" }
									props.conn.send(JSON.stringify(mess))
								}
							}}>Raise</button>
						</div>
					) : (
						<div className="flex justify-end h-14 mb-4">
							<button className="mr-4 py-2 px-10 rounded text-white font-bold bg-gray-800">Fold</button>
							<button className="mr-4 py-2 px-10 rounded text-white font-bold bg-gray-800">Check</button>
							<button className="mr-4 py-2 px-10 rounded text-white font-bold bg-gray-800">Call</button>
							<button className="mr-4 py-2 px-10 rounded text-white font-bold bg-gray-800">Raise</button>
						</div>
					)
				}
			</div>
		</div>
	)
}

export default Table;
