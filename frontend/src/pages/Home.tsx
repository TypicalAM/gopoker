import React from 'react'
import Sidebar from '../components/Sidebar'

function Home() {
	return (
		<div id="wrapper" className="p-5 bg-white dark:bg-gray-900 antialiased h-screen">
			<Sidebar />
			<div className="p-4 sm:ml-64">
				<h1 className="text-2xl font-bold text-gray-900 dark:text-gray-100">This is the home page of the application, if you wish to join a game, go the <a href="/game/queue" className="text-teal-500"> queue </a> </h1>
			</div>
		</div>
	)
}

export default Home
