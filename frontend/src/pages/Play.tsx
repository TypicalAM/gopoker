import React from 'react';
import Gameplay from '../components/Gameplay';
import Sidebar from '../components/Sidebar';

function Play() {
	return (
		<div id="wrapper" className="bg-white dark:bg-gray-900 antialiased h-screen">
			<Sidebar />
			<div className="h-full sm:ml-64">
				<Gameplay />
			</div>
		</div>
	)
}

export default Play;
