import React, { FormEvent } from 'react';
import { useNavigate } from 'react-router-dom';
import ErrorPopup from './Popup';

function LoginForm() {
	const [showPopup, setShowPopup] = React.useState(false);
	const [errorMessage, setErrorMessage] = React.useState('');
	const [username, setUsername] = React.useState('');
	const [password, setPassword] = React.useState('');

	const navigate = useNavigate();

	const handleClosePopup = () => {
		setShowPopup(false);
		setErrorMessage('');
	}

	const handleSubmit = (event: FormEvent<HTMLFormElement>) => {
		event.preventDefault();

		fetch(process.env.REACT_APP_API_URL + '/api/login', {
			method: 'POST',
			credentials: 'include',
			body: JSON.stringify({ username, password }),
			headers: {
				'Content-Type': 'application/json',
			},
		})
			.then(response => response.json())
			.then(data => {
				if (data.error) {
					setShowPopup(true);
					setErrorMessage(data.error);
				} else {
					setShowPopup(false);
					setErrorMessage('');
					localStorage.setItem('isAuthenticated', 'true');
					navigate('/');
				}
			})
	}

	return (
		<div className="flex flex-col items-center justify-center px-6 py-8 mx-auto md:h-screen lg:py-0">

			{showPopup && (
				< ErrorPopup
					message="There has been an error logging in."
					error={errorMessage}
					onClose={handleClosePopup}
				/>
			)}

			<a href="/" className="flex items-center mb-6 text-2xl font-bold text-gray-900 dark:text-gray-100">
				<img className="w-8 h-8 mr-2" src="https://tailwindui.com/img/logos/workflow-mark-indigo-600.svg" alt="Workflow" />
				Gopoker
			</a>
			<div className="w-screen bg-white rounded-lg shadow dark:border md:mt-0 sm:max-w-md xl:p-0 dark:bg-gray-800 dark:border-gray-700">
				<div className="p-6 space-y-4 md:space-y-6 sm:p-8">
					<h1 className="text-xl font-bold leading-tight text-gray-900 md:text-2xl dark:text-white">
						Let's log you in!
					</h1>
					<form onSubmit={handleSubmit} className="space-y-4 md:space-y-6">
						<div>
							<label htmlFor="username" className="block mb-2 text-sm font-medium text-gray-900 dark:text-white">Your username</label>
							<input type="text" value={username} onChange={e => setUsername(e.target.value)} name="username" id="username" className="bg-gray-50 border border-gray-300 text-gray-900 sm:text-sm rounded-lg focus:ring-primary-600 focus:border-primary-600 block w-full p-2.5 dark:bg-gray-700 dark:border-gray-600 dark:placeholder-gray-400 dark:text-white dark:focus:ring-blue-500 dark:focus:border-blue-500" placeholder="A Tiny Rat" required></input>
						</div>
						<div>
							<label htmlFor="password" className="block mb-2 text-sm font-medium text-gray-900 dark:text-white">Password</label>
							<input type="password" value={password} onChange={e => setPassword(e.target.value)} name="password" id="password" placeholder="••••••••" className="bg-gray-50 border border-gray-300 text-gray-900 sm:text-sm rounded-lg focus:ring-primary-600 focus:border-primary-600 block w-full p-2.5 dark:bg-gray-700 dark:border-gray-600 dark:placeholder-gray-400 dark:text-white dark:focus:ring-blue-500 dark:focus:border-blue-500" required></input>
						</div>
						<button type="submit" className="w-full bg-gray-200 text-gray-900 dark:text-white bg-primary-600 hover:bg-primary-700 focus:ring-4 focus:outline-none focus:ring-primary-300 font-medium rounded-lg text-sm px-5 py-2.5 text-center dark:bg-gray-600 dark:hover:bg-primary-700 dark:focus:ring-primary-800">Log in</button>
						<p className="text-sm font-light text-gray-500 dark:text-gray-400">
							Do you wish to? <a href="/register" className="font-medium text-primary-600 hover:underline dark:text-primary-500">Sign up</a> instead?
						</p>
					</form>
				</div>
			</div>
		</div>
	)
}

export default LoginForm;
