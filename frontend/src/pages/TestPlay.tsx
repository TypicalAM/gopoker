import React from 'react';
import { DefaultGameState } from '../components/GameState';
import Sidebar from '../components/Sidebar';
import TestTable from '../components/TestTable';

function TestPlay() {
	return (
		<div id="wrapper" className="bg-white dark:bg-gray-900 antialiased h-screen">
			<Sidebar />
			<div className="flex flex-col h-screen sm:ml-64">

				<div className="flex top-0 items-center h-16 left-0 bg-gray-50 dark:bg-gray-700 overflow-auto">
					<h1 className="ml-5 left-0 text-2xl font-bold text-gray-900 dark:text-gray-100">
						Example name - Connected
					</h1>
				</div>

				<div className="flex-grow items-center justify-center">
					<TestTable {...DefaultGameState} />
				</div>

			</div>
		</div>
	)
}

export default TestPlay;
