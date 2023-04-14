import React, { useEffect } from 'react';
import { GameState } from './GameState';

function Table(props: GameState | null) {
	useEffect(() => {
		// Get the active game
		console.log(props);
	}, [props]);

	return (
		<div>
			<h1 className="text-2xl font-bold text-gray-300">Table</h1>
		</div>
	);
}

export default Table;
