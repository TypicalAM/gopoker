import React from 'react';
import ProfileCard from '../components/ProfileCard';
import Sidebar from '../components/Sidebar';

function Profile() {
	return (
		<div id="wrapper" className="bg-white dark:bg-gray-900 antialiased h-screen">
			<Sidebar />
			<div className="flex flex-col justify-center items-center h-screen sm:ml-64">
				<ProfileCard />
			</div>
		</div>
	)
}

export default Profile;
