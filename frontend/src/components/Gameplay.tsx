import React, { useEffect, useRef, useState } from 'react';
import { GameState, DefaultGameState } from './GameState';
import Table from './Table';

enum msgType {
	State = 'state',
	Error = 'error',
	Input = 'input',
}

interface GameMessage {
	type: msgType;
	data: string;
}

function Gameplay() {
	const ws = useRef<WebSocket | null>(null);
	const [gameName, setGameName] = useState('');
	const [statusMessage, setStatusMessage] = useState('');
	const [gameState, setGameState] = useState<GameState>(DefaultGameState);
	const id = require('uuid-readable');

	const handleMessage = (data: string) => {
		let gameMessage: GameMessage;
		try {
			gameMessage = JSON.parse(data) as GameMessage;
		} catch (e) {
			console.error('Invalid message received from websocket');
			return;
		}

		switch (gameMessage.type) {
			case msgType.State:
				let newGameState: GameState;
				try {
					newGameState = JSON.parse(gameMessage.data);
				} catch (e) {
					console.error('Invalid state message received from websocket');
					return;
				}

				setGameState(newGameState);
				break;

			case msgType.Error:
				// TODO: Handle error
				console.log("Received an error from the server");
				console.log(gameMessage.data);
				break

			case msgType.Input:
				// TODO: Handle input
				console.log("Received an input message from the server");
				console.log(gameMessage.data);
				break
		}
	}

	useEffect(() => {
		// Get the active game
		const activeGame = localStorage.getItem('activeGame');
		if (!activeGame) return;
		setGameName(id.short(activeGame));

		// Only connect if we haven't already
		if (ws.current) return;

		setStatusMessage('Connecting to websocket...');

		// Connect to the websocket
		const connString = 'ws://localhost:8080/api/game/id/' + activeGame;
		ws.current = new WebSocket(connString);

		ws.current.onopen = () => {
			setStatusMessage('Connected');
		};

		ws.current.onmessage = (event) => {
			handleMessage(event.data);
		};

		ws.current.onerror = () => {
			setStatusMessage('Error connecting to websocket');
			localStorage.removeItem('activeGame');
			setTimeout(() => {
				window.location.replace('/');
			}, 5000);
		}
	}, []);


	return (
		<div>
			<div className="flex top-0 items-center h-16 left-0 bg-gray-50 dark:bg-gray-700 overflow-auto">
				<h1 className="ml-5 left-0 text-2xl font-bold text-gray-900 dark:text-gray-100">
					{gameName} - {statusMessage}
				</h1>
			</div>
			<div className="flex flex-col items-center justify-center h-screen">
				<div>
					<Table {...gameState} />
				</div>
				<button className="mt-4 bg-gray-700 p-3 rounded-xl text-gray-100" onClick={() => {
					if (ws.current) {
						let type = 'action';
						let data = 'call';
						let payload = {
							type: type,
							data: data
						}
						ws.current.send(JSON.stringify(payload));
					}
				}}>Send Message</button>
			</div>
		</div>
	)
}

export default Gameplay;

