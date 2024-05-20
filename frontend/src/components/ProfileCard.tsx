import React, { useEffect } from 'react';
import ErrorPopup from './Popup';

interface ProfileData {
	user: {
		username: string;
		profile: {
			display_name: string;
			image_url: string;
		}
	}
}


function ProfileCard() {
	const [files, setFiles] = React.useState<FileList | null>(null);
	const [buttonText, setButtonText] = React.useState<string>("Edit Profile");

	const [newDisplayName, setNewDisplayName] = React.useState<string | null>(null);

	const [username, setUsername] = React.useState<string | null>(null);
	const [displayName, setDisplayName] = React.useState<string | null>(null);
	const [image, setImage] = React.useState<string | null>(null);

	const [editMode, setEditMode] = React.useState<boolean>(false);
	const [showPopup, setShowPopup] = React.useState(false);
	const [errorMessage, setErrorMessage] = React.useState('');

	const handleClosePopup = () => {
		setShowPopup(false);
		setErrorMessage('');
	}

	const fetchProfile = async () => {
		let resp = await fetch(process.env.REACT_APP_API_URL + '/api/profile', {
			method: 'GET', credentials: 'include', headers: { 'Content-Type': 'application/json' }
		})

		if (resp.status === 401) {
			localStorage.setItem('isAuthenticated', 'false');
			window.location.replace('/login');
			return
		}

		let data = await resp.json() as ProfileData
		setUsername(data.user.username);
		setDisplayName(data.user.profile.display_name);

		console.log(data.user.profile.image_url)
		if (data.user.profile.image_url.startsWith("https")) {
			setImage(`${data.user.profile.image_url}`);
			return
		}

		if (!process.env.REACT_APP_API_URL) {
			setImage(`${data.user.profile.image_url}?date=${Date.now()}`);
			return
		}

		setImage(`${process.env.REACT_APP_API_URL}${data.user.profile.image_url}?date=${Date.now()}`);
	}

	const handleSubmit = () => {
		if (displayName === newDisplayName) return;
		if (files == null || files.length === 0) return;

		const reader = new FileReader();
		reader.readAsDataURL(files![0]);
		reader.onload = async () => {
			const body = JSON.stringify({ display_name: newDisplayName, image_data: reader.result })
			let resp = await fetch(process.env.REACT_APP_API_URL + '/api/profile', {
				method: 'PUT', credentials: 'include', headers: { 'Content-Type': 'application/json' }, body: body,
			})

			if (resp.status !== 200) {
				if (resp.status === 401) {
					localStorage.setItem('isAuthenticated', 'false');
					window.location.replace('/login');
					return
				}

				let data = await resp.json()
				if (data.error) {
					setShowPopup(true)
					setErrorMessage(data.error)
				}

				return
			}

			window.location.replace("/profile");
		}
	}

	const toggleEditMode = (e: React.MouseEvent<HTMLButtonElement, MouseEvent>) => {
		e.preventDefault();

		if (editMode) {
			handleSubmit();
		}

		setEditMode(!editMode);
		setButtonText(editMode ? "Edit Profile" : "Save Profile");
	}

	useEffect(() => {
		fetchProfile();
	}, []);

	return (
		<div className="max-w-sm bg-white border border-gray-200 rounded-lg shadow dark:bg-gray-800 dark:border-gray-700">
			{showPopup && (
				<ErrorPopup
					message="There has been an error updating your profile."
					error={errorMessage}
					onClose={handleClosePopup}
				/>
			)}

			<a href="#">
				{image && <img className="rounded-t-lg" src={image} alt="avatar" />}
			</a>
			<div className="p-5">
				{(editMode) ?
					(
						<div className="flex items-center mb-4">
							<h1 className="text-2xl font-bold text-gray-800 dark:text-white md:text-3xl hover:underline focus:underline mr-5">Hello </h1>
							<input type="text" id="first_name" className="bg-gray-50 border border-gray-300 text-gray-900 text-md rounded-lg focus:ring-blue-500 focus:border-blue-500 block w-full p-2.5 dark:bg-gray-700 dark:border-gray-600 dark:placeholder-gray-400 dark:text-white dark:focus:ring-blue-500 dark:focus:border-blue-500" placeholder={displayName || "name"} onChange={(e) => { setNewDisplayName(e.target.value) }}></input>
						</div>
					) : (
						<h1 className="text-2xl font-bold text-gray-800 dark:text-white md:text-3xl hover:underline focus:underline">Hello {displayName}</h1>
					)
				}

				<p className="mb-3 font-normal text-gray-700 dark:text-gray-400">Your username is {username}</p>
				{editMode && (
					<div>
						<input className="mb-5 block w-full text-sm text-gray-900 border border-gray-300 rounded-lg cursor-pointer bg-gray-50 dark:text-gray-400 focus:outline-none dark:bg-gray-700 dark:border-gray-600 dark:placeholder-gray-400" type="file" onChange={(e) => setFiles(e.target.files)} ></input>
					</div>
				)}
				<button className="inline-flex items-center px-3 py-2 text-sm font-medium text-center text-white bg-blue-700 rounded-lg hover:bg-blue-800 focus:ring-4 focus:outline-none focus:ring-blue-300 dark:bg-blue-600 dark:hover:bg-blue-700 dark:focus:ring-blue-800" onClick={toggleEditMode}>
					{buttonText}
					<svg aria-hidden="true" className="w-4 h-4 ml-2 -mr-1" fill="currentColor" viewBox="0 0 20 20" xmlns="http://www.w3.org/2000/svg"><path fillRule="evenodd" d="M10.293 3.293a1 1 0 011.414 0l6 6a1 1 0 010 1.414l-6 6a1 1 0 01-1.414-1.414L14.586 11H3a1 1 0 110-2h11.586l-4.293-4.293a1 1 0 010-1.414z" clipRule="evenodd"></path></svg>
				</button>
			</div>
		</div>
	);
}

export default ProfileCard;
