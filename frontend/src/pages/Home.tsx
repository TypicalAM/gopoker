import React from 'react'
import TableComponent from '../components/TableComponent'
import Sidebar from '../components/Sidebar'

function Home() {
	return (
		<div id="wrapper" className="p-5 bg-white dark:bg-gray-900 antialiased h-screen">
			<Sidebar />
			<div className="p-4 sm:ml-64">
				<TableComponent />
			</div>
		</div>
	)
}

export default Home
