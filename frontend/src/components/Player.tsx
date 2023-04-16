import React, { useEffect } from 'react';
import { Player } from "./GameState";
import PlayingCard from './Card';

interface PlayerCardProps {
	Value: Player;
	Active: boolean;
	IsMe: boolean;
	HasWon: boolean;
	GameOver: boolean;
}

function PlayerCard(props: PlayerCardProps) {
	const [actionDescription, setActionDescription] = React.useState("Waiting");
	const [background, setBackground] = React.useState("bg-gray-800");

	useEffect(() => {
		if (props.HasWon) {
			setBackground("bg-gradient-to-t from-emerald-300 to-gray-800");
		} else if (!props.GameOver && props.IsMe && props.Active) {
			setBackground("bg-gradient-to-t from-violet-300 to-gray-800");
		} else if (!props.GameOver && props.Active) {
			setBackground("bg-gradient-to-t from-amber-300 to-gray-800");
		} else {
			setBackground("bg-gray-800");
		}

		if (!props.HasWon) {
		switch (props.Value.Action) {
			case "none":
				setActionDescription("Waiting");
				break;
			case "fold":
				setActionDescription("Folded");
				break;
			case "call":
				setActionDescription("Called");
				break;
			case "raise":
				setActionDescription("Raised");
				break;
			case "check":
				setActionDescription("Checked");
				break;
		}
		} else {
			setActionDescription("Winner!");
		}
	}, [props])

	return (
		<div className={`flex flex-col h-1/2 mx-4 p-4 rounded-xl justify-center items-center ${background}`}>
			<h1 className="font-bold text-2xl text-white mb-3">{props.Value.Name}</h1>

			<div className="flex flex-row justify-between w-full">
				<h1 className="font-bold text-l ml-5 text-white">{actionDescription}</h1>
				<h1 className="font-bold text-l mr-5 text-white">Bets <span className="ml-1 text-yellow-500">{props.Value.Bet}</span></h1>
			</div>

			<div className="flex space-x-4">
				<PlayingCard Value={props.Value.HoleCards[0]} />
				<PlayingCard Value={props.Value.HoleCards[1]} />
			</div>
		</div>
	)
}

export default PlayerCard;
