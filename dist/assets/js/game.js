window.onload = main;

function main() {
	let conn;
	let table = document.getElementById("table_body")
	let lastMessage = "";
	let winnerID = "";
	let winnerRank = "";
	let gameEnded = false;

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

		let row = table.insertRow();
		let cell = row.insertCell();

		parsedMsg = JSON.parse(msg);
		switch (parsedMsg.type) {
			case "end":
				if (gameEnded) break;
				console.log(parsedMsg.data);
				gameEnded = true;
				winnerID = Number(parsedMsg.data.split(":")[0]);
				winnerRank = parsedMsg.data.split(":")[1];
				let winnerDiv = document.getElementById("player_div").childNodes[winnerID + 1]; // Add 1 bcs of the community cards
				winnerDiv.classList.remove("alert-info");
				winnerDiv.classList.add("alert-success");
				console.log(winnerDiv);

				let text = winnerDiv.getElementsByClassName("left-half")[0].getElementsByTagName("p")[0];
				text.innerHTML = text.innerHTML.split(" (")[0] + " (Winner) - " + winnerRank + "";
				cell.innerHTML = "<b>Game ended!</b>";
				break;

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
				if (gameEnded) return;
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
		let titleDiv = document.getElementById("title");
		titleDiv.innerHTML = "Game of Texas Hold'em (Pot: " + stateData.TotalBets + ")"
		let allPlayersDiv = document.getElementById("player_div");
		allPlayersDiv.innerHTML = "";

		if (stateData.CommunityCards.length != 0) {
			let communityCardsDiv = document.createElement("div");
			communityCardsDiv.classList.add("alert");
			communityCardsDiv.classList.add("alert-info");
			communityCardsDiv.classList.add("community-cards");

			for (i = 0; i < stateData.CommunityCards.length; i++) {
				let card = document.createElement("img");
				card.classList.add("playingcard");
				card.src = "/assets/images/cards/" + stateData.CommunityCards[i] + ".png"
				communityCardsDiv.appendChild(card);
			}

			allPlayersDiv.appendChild(communityCardsDiv);
		}

		for (i = 0; i < stateData.Usernames.length; i++) {
			let playerDiv = document.createElement("div");
			playerDiv.classList.add("alert");
			playerDiv.classList.add(stateData.Turn == i ? "alert-success" : "alert-info");
			playerDiv.classList.add("player");

			let leftHalf = document.createElement("div");
			leftHalf.classList.add("left-half");
			let usernameText = document.createElement("p");
			usernameText.innerHTML = stateData.Usernames[i] + " (" + stateData.Assets[i] + " coins)";
			leftHalf.appendChild(usernameText);
			playerDiv.appendChild(leftHalf);

			if (stateData.Turn == i && !gameEnded && stateData.Hands[i][0] != undefined) {
				// Check if this is us, that is the cards are not hidden
				let centerHalf = document.createElement("div");
				centerHalf.classList.add("center-half");

				// Create four buttons for the player to choose from
				let callButton = document.createElement("button");
				callButton.innerHTML = "Call";
				callButton.classList.add("center-half-button");
				callButton.onclick = function() {
					conn.send(JSON.stringify({ type: "action", data: JSON.stringify({ type: "call", data: "" }) }));
				}
				centerHalf.appendChild(callButton);

				let foldButton = document.createElement("button");
				foldButton.innerHTML = "Fold";
				foldButton.classList.add("center-half-button");
				foldButton.onclick = function() {
					conn.send(JSON.stringify({ type: "action", data: JSON.stringify({ type: "fold", data: "" }) }));
				}
				centerHalf.appendChild(foldButton);

				let raiseButton = document.createElement("button");
				raiseButton.innerHTML = "Raise";
				raiseButton.classList.add("center-half-button");
				raiseButton.onclick = function() {
					let raiseAmount = prompt("How much would you like to raise by?");
					conn.send(JSON.stringify({ type: "action", data: JSON.stringify({ type: "raise", data: raiseAmount }) }));
				}
				centerHalf.appendChild(raiseButton);

				if (stateData.Round != 0) {
					let checkButton = document.createElement("button");
					checkButton.innerHTML = "Check";
					checkButton.classList.add("center-half-button");
					checkButton.onclick = function() {
						conn.send(JSON.stringify({ type: "action", data: JSON.stringify({ type: "check", data: "" }) }));
					}
					centerHalf.appendChild(checkButton);
				}

				playerDiv.appendChild(centerHalf);
			}

			let rightHalf = document.createElement("div");
			rightHalf.classList.add("right-half");

			let action;
			switch (stateData.Actions[i]) {
				case 0:
					action = "No action yet";
					break;
				case 1:
					action = "Folded";
					break;
				case 2:
					action = "Checked";
					break;
				case 3:
					action = "Called";
					break;
				case 4:
					action = "Raised";
					break;
			}

			let currentBidText = document.createElement("p");
			currentBidText.innerHTML = action + " - bidding " + stateData.Bets[i];
			rightHalf.appendChild(currentBidText);

			let firstCard = document.createElement("img");
			firstCard.classList.add("playingcard");
			if (stateData.Hands[i][0] == undefined) {
				firstCard.src = "/assets/images/cards/back.png"
			} else {
				firstCard.src = "/assets/images/cards/" + stateData.Hands[i][0] + ".png"
			}
			rightHalf.appendChild(firstCard);

			let secondCard = document.createElement("img");
			secondCard.classList.add("playingcard");
			if (stateData.Hands[i][1] == undefined) {
				secondCard.src = "/assets/images/cards/back.png"
			} else {
				secondCard.src = "/assets/images/cards/" + stateData.Hands[i][1] + ".png"
			}
			rightHalf.appendChild(secondCard);

			playerDiv.appendChild(rightHalf);

			allPlayersDiv.appendChild(playerDiv);
		}
	}
}


