import React from 'react';
import { createBrowserRouter, RouterProvider } from 'react-router-dom';
import Home from './pages/Home';
import Register from './pages/Register';
import Login from './pages/Login';
import Logout from './pages/Logout';
import Queue from './pages/Queue';
import Play from './pages/Play';
import TestPlay from './pages/TestPlay';

const router = createBrowserRouter([
	{
		path: '/',
		element: <Home />,
	},
	{
		path: '/register',
		element: <Register />,
	},
	{
		path: '/login',
		element: <Login />,
	},
	{
		path: '/logout',
		element: <Logout />,
	},
	{
		path: '/game/queue',
		element: <Queue />,
	},
	{
		path: '/game/play',
		element: <Play />,
	},
	{
		path: '/game/testPlay',
		element: <TestPlay />,
	}
]);

function App() {
	return (
		<RouterProvider router={router} />
	);
}

export default App;
