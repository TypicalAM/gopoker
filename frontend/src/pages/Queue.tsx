import React from 'react';
import JoinQueue from '../components/JoinQueue';
import Sidebar from '../components/Sidebar';

function Queue() {
	return (
		<div id="wrapper" className="p-5 bg-white dark:bg-gray-900 antialiased h-screen">
			<Sidebar />
			<div className="p-4 h-full flex flex-col justify-center items-center sm:ml-64">
				<JoinQueue />
			</div>
		</div>
	);
}

export default Queue;
