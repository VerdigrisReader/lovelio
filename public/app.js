var Socket = new WebSocket("ws://localhost:3000/ws");

Socket.onopen = function (event) {
    Socket.send("HIYA")
}

Socket.onmessage = function(event) {
    message = event.data
    console.log(message)
    switch(message.type) {
        case "boardData":
            break;
        case "boardChange":
            break;
    }
}

function newBoard() {
    Socket.send({"messageType": "newBoard"})
}
