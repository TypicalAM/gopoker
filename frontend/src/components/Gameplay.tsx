import React from 'react';

function Gameplay() {
	const [gameID, setGameID] = React.useState('');
	const [gameName, setGameName] = React.useState('');
	const [error, setError] = React.useState(false);
	const [errorMsg, setErrorMsg] = React.useState('');
	const [messsages, setMessages] = React.useState([]);
	const id = require('uuid-readable');

	const getGameID = () => {
		const activeGameID = localStorage.getItem('activeGame');
		if (activeGameID) {
			setGameName(id.short(activeGameID));
			setGameID(activeGameID);
		} else {
			setGameName('No game');
			setError(true);
			setErrorMsg('No active game found. Please create a new game.');
		}
	}

	const establishConnection = () => { 
		const socket = new WebSocket('ws://localhost:8080');

		socket.onopen = (event) => {
			//setMessages((prevMessages) => [...prevMessages, 'Connected to server']);
		}

		socket.onmessage = (event) => {
			console.log(event);
		}
	}

	React.useEffect(() => {
		getGameID();
		establishConnection();
	}, [])

	return (
		<div>
			<h1 className="text-2xl font-bold text-gray-900 dark:text-gray-100">
				{gameName}
			</h1>
			{error && (
				<div className="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded relative" role="alert">
					<strong className="font-bold">Error!</strong>
					<span className="block sm:inline">{errorMsg}</span>
				</div>
			)}
		</div>
	)
}

export default Gameplay;

