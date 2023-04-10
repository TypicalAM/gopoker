import React from 'react';
import Gameplay from '../components/Gameplay';
import Sidebar from '../components/Sidebar';

function Play() {
	return (
		<div id="wrapper" className="p-5 bg-white dark:bg-gray-900 antialiased h-screen">
			<Sidebar />
			<div className="p-4 h-full flex flex-col justify-center items-center sm:ml-64">
				<Gameplay />
			</div>
		</div>
	)
}

export default Play;
