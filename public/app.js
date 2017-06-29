var Socket = new WebSocket("ws://localhost:3000/ws");

Socket.onopen = function (event) {
    Socket.send("HIYA")
}

Socket.onmessage = function(event) {
    console.log(event.data)
}
