window.onload = main;

function main() {
	let conn;
	let lastMessage = "";
	let secondToLastMessage = "";
	let table = document.getElementById("table_body")

	if (window["WebSocket"]) {
		let target = "ws://" + window.location.href.substring(7) + "/ws";
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
		console.log("Received message: " + msg);
		if (msg === lastMessage || msg === secondToLastMessage) {
			console.log("Ignoring duplicate message.");
			return;
		}

		secondToLastMessage = lastMessage;
		lastMessage = msg;

		let row = table.insertRow();
		let cell = row.insertCell();

		switch (msg.substring(0, 6)) {
			case 'action':
				let availableActions = msg.substring(7).split(',');
				cell.innerHTML = "<b>Available actions:</b>";
				for (let i = 0; i < availableActions.length; i++) {
					let action = availableActions[i];
					let button = document.createElement("button");
					button.innerHTML = action;
					button.onclick = function() { conn.send(action); };
					cell.appendChild(button);
				}
				break;

			case 'status':
				let statusMsg = msg.substring(7);
				cell.innerHTML = "<b>Status:</b> " + statusMsg;
				break;

			case 'uinput':
				let inputMsg = msg.substring(7);
				textInput = document.createElement("input");
				textInput.type = "text";
				textInput.placeholder = inputMsg;
				button = document.createElement("button");
				button.innerHTML = "Send";
				button.onclick = function() { conn.send("uinput:"+textInput.value); };
				cell.appendChild(textInput);
				cell.appendChild(button);
				break;

			default:
				cell.innerHTML = "Received a stupid message, ignoring it.";
				break;
		}
	}
}


