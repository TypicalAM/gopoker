import React from 'react';
import Gameplay from '../components/Gameplay';
import Sidebar from '../components/Sidebar';

function Play() {
	return (
		<div id="wrapper" className="bg-white dark:bg-gray-900 antialiased">
			<Sidebar />
			<div className="flex flex-col sm:ml-64 h-screen">
				<Gameplay />
			</div>
		</div>
	)
}

export default Play;
