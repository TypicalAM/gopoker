import React, { useEffect } from 'react';
import { GameState } from './GameState';
import PlayerCard from './Player';
import Card from './Card';

function TestTable(props: GameState) {
	const [myIndex, setMyIndex] = React.useState(-1);
	const [communityCards, setCommunityCards] = React.useState<string[]>([""]);

	useEffect(() => {
		for (let i = 0; i < props.Players.length; i++) {
			if (props.Players[i].HoleCards.length > 0) {
				console.log("My index is " + i);
				setMyIndex(i);
			}
		}

		// Pad the community cards so there is 5 strings
		let cards = props.CommunityCards;
		if (!cards) {
			cards = [];
		}

		while (cards.length < 5) {
			cards.push("");
		}

		setCommunityCards(cards);
	}, [props])

	return (
		<div className="w-full p-4 h-max">
			<div className="flex flex-col bg-gradient-to-br from-gray-800 to-gray-700 h-full w-full rounded-xl space-y-10 pb-5">
				<div className="top-0 left-0 pl-4 pt-2">
					<h1 className="text-xl font-bold text-gray-900 dark:text-gray-100">
						Pot <span className="ml-2 text-red-500">1000</span>
					</h1>
				</div>

				<div className="flex flex-col items-center space-y-4">
					<h1 className="text-2xl font-bold text-gray-900 dark:text-gray-100">
						Current Bet <span className="ml-2 text-emerald-500">{props.ActiveBet}</span>
					</h1>
					<div className="flex justify-center space-x-4 bg-gray-800 rounded-xl p-3">
						{communityCards.map((card, _) => {
							if (!card) {
								return <Card {...{ Value: null, IsCommunity: true }} />
							} else {
								return <Card {...{ Value: card, IsCommunity: true }} />
							}
						})
						}
					</div>
				</div>

				<div className="flex flex-row justify-center items-center">

					<PlayerCard {...{ Value: props.Players[0], Active: 0 === props.CurrentPlayer, IsMe: myIndex === 0 }} />
					<PlayerCard {...{ Value: props.Players[1], Active: 1 === props.CurrentPlayer, IsMe: myIndex === 1 }} />
					<PlayerCard {...{ Value: props.Players[2], Active: 2 === props.CurrentPlayer, IsMe: myIndex === 2 }} />

				</div>

				{
					props.CurrentPlayer === myIndex ? (
						<div className="flex justify-end h-14 mb-4">
							<button className="mr-4 py-2 px-10 rounded text-white font-bold bg-gradient-to-br from-red-400 to-red-500 hover:bg-gradient-to-br hover:from-red-500 hover:to-red-500">Fold</button>
							<button className="mr-4 py-2 px-10 rounded text-white font-bold bg-gray-800 hover:bg-gray-900">Check</button>
							<button className="mr-4 py-2 px-10 rounded text-white font-bold bg-gray-800 hover:bg-gray-900">Call</button>
							<button className="mr-4 py-2 px-10 rounded text-white font-bold bg-gradient-to-br from-emerald-400 to-emerald-500 hover:bg-gradient-to-br hover:from-emerald-500 hover:to-emerald-500">Raise</button>
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

export default TestTable;
