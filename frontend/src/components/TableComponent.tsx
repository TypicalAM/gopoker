import React from 'react';

function TableComponent() {
	// Check if the user is authenticated
	let isAuthenticated = false;

	let message;
	// Send a GET request to the API and check if the user is authenticated
	fetch('http://localhost:8080/check', {
		method: 'POST',
		headers: {
			'Content-Type': 'application/json',
		}
	}).then(response => response.json()).then(data => {
		console.log(data)
		message = data;
	});

	return (
		<h1 className="text-3xl dark:text-red-300"> Here we will have the table!, isAuthenticated: {isAuthenticated} </h1>
	)
}

export default TableComponent;
