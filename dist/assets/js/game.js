window.onload = main;

function main() {
	let conn;
	let table = document.getElementById("table_body")
	let lastMessage = "";

	if (window["WebSocket"]) {
		let target = "ws://" + window.location.href.substring(7) + "ws";
		console.log("Connecting to " + target);
		conn = new WebSocket(target);
		conn.onmessage = onmessage;
		conn.onclose = close;
	} else {
		let row = table.insertRow();
		let cell = row.insertCell();
		cell.innerHTML = "<b>Your browser does not support WebSockets.</b>";
	}

	function close(evt) {
		console.log("Closing connection - ", evt);
		let row = table.insertRow();
		let cell = row.insertCell();
		cell.innerHTML = "<b>Connection closed.</b>";
	}

	function onmessage(evt) {
		let messages = evt.data.split('\n');
		for (let i = 0; i < messages.length; i++)	handleMessage(messages[i]);
	}

	function handleMessage(msg) {
		if (msg == lastMessage) {
			console.log("Duplicate message received, ignoring.");
			return;
		}

		lastMessage = msg;

		console.log("Received message: " + msg);

		let row = table.insertRow();
		let cell = row.insertCell();

		parsedMsg = JSON.parse(msg);
		switch (parsedMsg.type) {
			case "status":
				cell.innerHTML = "<b>Status: " + parsedMsg.data + "</b>";
				break;

			case "input":
				dataSplit = parsedMsg.data.split(":");
				for (let i = 0; i < dataSplit.length; i++) {
					let button = document.createElement("button");
					button.innerHTML = dataSplit[i];
					button.onclick = function() {
						let messageObj;
						if (dataSplit[i] == "raise") {
							let raiseAmount = prompt("How much would you like to raise by?");
							messageObj = { type: "action", data: JSON.stringify({ type: dataSplit[i], data: raiseAmount }) };
						} else {
							messageObj = { type: "action", data: JSON.stringify({ type: dataSplit[i], data: "" }) };
						}

						conn.send(JSON.stringify(messageObj));
					}
					cell.appendChild(button);
				}
				break;

			case "start":
				cell.innerHTML = "<b>Game started!</b>";
				break;

			case "state":
				console.log("We received a new state message!");
				updateState(JSON.parse(parsedMsg.data))
				break;

			case "players":
				console.log("We received a new players message!");
				console.log(JSON.parse(parsedMsg.data));
				break;

			default:
				cell.innerHTML = "<b>Unknown message type: " + parsedMsg.type + "</b>";
				break;
		}
	}

	function updateState(stateData) {
		let allPlayersDiv = document.getElementById("player_div");
		allPlayersDiv.innerHTML = "";

		console.log("Hello, from the start of the function");
		console.log(stateData);
		for (i = 0; i < stateData.Usernames.length; i++) {
			console.log("Hello from the inside, the user is " + stateData.Usernames[i]);
			let playerDiv = document.createElement("div");
			playerDiv.classList.add("alert");
			playerDiv.classList.add(stateData.Turn == i ? "alert-success" : "alert-info");
			playerDiv.classList.add("player");

			let leftHalf = document.createElement("div");
			leftHalf.classList.add("left-half");
			let usernameText = document.createElement("p");
			usernameText.innerHTML = stateData.Usernames[i];
			leftHalf.appendChild(usernameText);
			playerDiv.appendChild(leftHalf);

			let rightHalf = document.createElement("div");
			rightHalf.classList.add("right-half");
			let firstCard = document.createElement("img");
			firstCard.classList.add("playingcard");
			if (stateData.Hands[i][0] == undefined) {
				firstCard.src = "/assets/images/cards/As.png"
			} else {
				firstCard.src = "/assets/images/cards/" + stateData.Hands[i][0] + ".png"
			}
			console.log(stateData.Hands[i][0]);
			rightHalf.appendChild(firstCard);

			let secondCard = document.createElement("img");
			secondCard.classList.add("playingcard");
			if (stateData.Hands[i][1] == undefined) {
				secondCard.src = "/assets/images/cards/As.png"
			} else {
				secondCard.src = "/assets/images/cards/" + stateData.Hands[i][1] + ".png"
			}
			console.log(stateData.Hands[i][1]);
			rightHalf.appendChild(secondCard);

			playerDiv.appendChild(rightHalf);

			allPlayersDiv.appendChild(playerDiv);
		}
	}
}


