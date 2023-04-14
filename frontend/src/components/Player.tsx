import React, { useEffect } from 'react';
import { Player } from "./GameState";
import Card from './Card';

interface PlayerCardProps {
	Value: Player;
	Active: boolean;
	IsMe: boolean;
}

function PlayerCard(props: PlayerCardProps) {
	const [actionDescription, setActionDescription] = React.useState("Waiting");
	const [background, setBackground] = React.useState("bg-gray-800");

	useEffect(() => {
		if (props.IsMe && props.Active) {
			setBackground("bg-gradient-to-t from-violet-300 to-gray-800");
		} else if (props.Active) {
			setBackground("bg-gradient-to-t from-amber-300 to-gray-800");
		} else {
			setBackground("bg-gray-800");
		}

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
	}, [props])

	return (
		<div className={`flex flex-col h-1/2 mx-4 p-4 rounded-xl justify-center items-center ${background}`}>
			<h1 className="font-bold text-2xl text-white mb-3">{props.Value.Name}</h1>
			<h1 className="font-bold text-m text-white mb-5">{actionDescription}</h1>
			<div className="flex space-x-4">
				<Card Value={props.Value.HoleCards[0]} />
				<Card Value={props.Value.HoleCards[0]} />
			</div>
		</div>
	)
}

export default PlayerCard;
