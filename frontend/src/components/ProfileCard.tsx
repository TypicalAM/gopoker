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

interface UploadBody {
	display_name: string;
	image_data: string | ArrayBuffer | null;
	password: string;
}


function ProfileCard() {
	const [files, setFiles] = React.useState<FileList | null>(null);
	const [buttonText, setButtonText] = React.useState<string>("Edit Profile");

	const [newDisplayName, setNewDisplayName] = React.useState<string | null>(null);
	const [password, setPassword] = React.useState<string | null>(null);

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

	const sendChanged = async (body: UploadBody) => {
		console.log(`Sending this body:`, body)
		let resp = await fetch(process.env.REACT_APP_API_URL + '/api/profile', {
			method: 'PUT', credentials: 'include', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(body),
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
		return;
	};

	const handleSubmit = async () => {
		let body = {} as UploadBody;

		if (newDisplayName !== null && displayName !== newDisplayName) body.display_name = newDisplayName
		if (password !== null && password.length !== 0) body.password = password;
		if (Object.keys(body).length !== 0 && (files === null || files.length === 0)) {
			console.log("im doing this", body)
			await sendChanged(body)
			return
		}

		if (files === null || files.length === 0) return;
		const reader = new FileReader();
		reader.readAsDataURL(files![0]);
		reader.onloadend = async () => {
			console.log("halo!")
			body.image_data = reader.result;
			console.log(reader.result)
			await sendChanged(body)
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
		<div className="w-full max-w-sm bg-white border border-gray-200 rounded-lg shadow dark:bg-gray-800 dark:border-gray-700">
			{showPopup && (
				<ErrorPopup
					message="There has been an error updating your profile."
					error={errorMessage}
					onClose={handleClosePopup}
				/>
			)}

			<a href="#">
				{image && <img className="p-8 rounded-t-lg" src={image} alt="avatar" />}
			</a>
			<div className="px-5 pb-5">
				<a href="#">
					<h5 className="text-xl font-semibold tracking-tight text-gray-900 dark:text-white">Hello {username}, better known as {displayName}</h5>
				</a>
				{editMode && (
					<div className='mt-4'>
						<div>
							<label htmlFor="name" className="block mb-2 text-sm font-medium text-gray-900 dark:text-white">Display Name</label>
							<input type="name" name="name" id="name" className="bg-gray-50 border border-gray-300 text-gray-900 text-sm rounded-lg focus:ring-blue-500 focus:border-blue-500 block w-full p-2.5 dark:bg-gray-600 dark:border-gray-500 dark:placeholder-gray-400 dark:text-white" onChange={(e) => { setNewDisplayName(e.target.value) }} placeholder={displayName!} />
						</div>
						<div>
							<label htmlFor="password" className="block mb-2 text-sm font-medium text-gray-900 dark:text-white">Your password</label>
							<input type="password" name="password" id="password" placeholder="••••••••" className="bg-gray-50 border border-gray-300 text-gray-900 text-sm rounded-lg focus:ring-blue-500 focus:border-blue-500 block w-full p-2.5 dark:bg-gray-600 dark:border-gray-500 dark:placeholder-gray-400 dark:text-white" onChange={(e) => { setPassword(e.target.value) }} />
						</div>
						<div>
							<label className="block mb-2 text-sm font-medium text-gray-900 dark:text-white" htmlFor="avatar">Upload avatar</label>
							<input className="block w-full text-sm text-gray-900 border border-gray-300 rounded-lg cursor-pointer bg-gray-50 dark:text-gray-400 focus:outline-none dark:bg-gray-700 dark:border-gray-600 dark:placeholder-gray-400" id="avatar" type="file" onChange={(e) => setFiles(e.target.files)} />
						</div>
					</div>
				)}
				<div className="flex items-center justify-center">
					{editMode ? (
						<button className="text-white bg-blue-700 hover:bg-blue-800 focus:ring-4 focus:outline-none focus:ring-blue-300 font-medium rounded-lg text-sm mt-5 px-5 py-2.5 text-center dark:bg-blue-600 dark:hover:bg-blue-700 dark:focus:ring-blue-800" onClick={handleSubmit}>{buttonText}</button>

					) : (
						<button className="text-white bg-blue-700 hover:bg-blue-800 focus:ring-4 focus:outline-none focus:ring-blue-300 font-medium rounded-lg text-sm mt-5 px-5 py-2.5 text-center dark:bg-blue-600 dark:hover:bg-blue-700 dark:focus:ring-blue-800" onClick={toggleEditMode}>{buttonText}</button>
					)}
				</div>
			</div>
		</div >
	);
}

export default ProfileCard;
