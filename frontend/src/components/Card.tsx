import { useEffect, useState } from "react";
import heart from "../images/heart.png";
import clubs from "../images/clubs.png";
import spade from "../images/spade.png";
import diamond from "../images/diamond.png";
import question from "../images/question-mark.png";

interface PlayingCardProps {
	Value: string | null;
}

function PlayingCard(props: PlayingCardProps) {
	const [suit, setSuit] = useState(question);
	const [value, setValue] = useState("");
	const [red, setRed] = useState(false);

	useEffect(() => {
		if (!props.Value) {
			setValue("");
			return;
		}

		switch (props.Value[1]) {
			case "h":
				setSuit(heart);
				setRed(true);
				break;
			case "s":
				setSuit(spade);
				setRed(false);
				break;
			case "d":
				setSuit(diamond);
				setRed(true);
				break;
			case "c":
				setSuit(clubs);
				setRed(false);
				break;
		}

		if (props.Value[0] === 'T') setValue("10");
		else setValue(props.Value[0]);
	}, [props]);

	return (
		<div className="bg-gray-600 dark:bg-slate-700 h-32 w-24 rounded-xl border-2 border-dashed border-gray-500 m-4">
			{value ? (
				<div className="relative h-full w-full">
					<img alt="Card suit" src={suit} className="font-bold top-0 left-0 ml-3 mt-3 h-8 w-9"></img>
					{red ? (
						<h1 className="font-bold absolute mb-5 mr-4 bottom-0 right-0 text-4xl text-red-500 dark:text-red-500">{value}</h1>
					) : (
						<h1 className="font-bold absolute mb-5 mr-4 bottom-0 right-0 text-4xl text-gray-300 dark:text-gray-100">{value}</h1>
					)}
				</div>
			) : (
				<div className="flex justify-center items-center h-full w-full">
					<img alt="Unknown card" src={question} className="h-16 w-16"></img>
				</div>
			)}
		</div>
	)
}

export default PlayingCard;

