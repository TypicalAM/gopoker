import React from 'react'
import Sidebar from '../components/Sidebar'

function Home() {
	const id = require('uuid-readable');

	const promptGame = () => {
		const roomName = prompt("Please enter the room id here", "8ff21897-c64b-4ea9-b726-b6d1344e20c4")
		if (roomName === null || roomName === "") {
			console.log("User closed the room name prompt")
			return
		}

		const uuidRegex = /^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$/
		if (uuidRegex.test(roomName)) {
			localStorage.setItem('activeGame', roomName);
			window.location.replace('/game/play');
		}
	}

	return (
		<div id="wrapper" className="p-5 bg-white dark:bg-gray-900 antialiased h-screen">
			<Sidebar />
			<div className="p-4 sm:ml-64">
				<h1 className="text-2xl font-bold text-gray-900 dark:text-gray-100">This is the home page of the application, if you wish to join a game, go the <a href="/game/queue" className="text-teal-500"> queue </a> or <a href="#" onClick={promptGame} className="text-teal-500">enter a game with a room name</a></h1>
			</div>
		</div>
	)
}

export default Home
