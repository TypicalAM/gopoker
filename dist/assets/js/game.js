window.onload = main;

function main() {
	let conn;
	let table = document.getElementById("table_body")

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
		console.log("Received message: " + msg);

		let row = table.insertRow();
		let cell = row.insertCell();

		parsedMsg = JSON.parse(msg);
		switch (parsedMsg.type) {
			case "status":
				cell.innerHTML = "<b>Status: " + parsedMsg.data + "</b>";
				break;

			case "input":
				input = document.createElement("input");
				input.type = "text";
				input.placeholder = parsedMsg.data;
				button = document.createElement("button");
				button.innerHTML = "Send";
				button.onclick = function() {
					messageObj = {type: "action", data: input.value};
					conn.send(JSON.stringify(messageObj));
					input.value = "";
				}
				cell.appendChild(input);
				cell.appendChild(button);
				break;

			case "start":
				cell.innerHTML = "<b>Game started!</b>";
				break;

			default:
				cell.innerHTML = "<b>Unknown message type: " + parsedMsg.type + "</b>";
				break;
		}
	}
}


