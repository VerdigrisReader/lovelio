var Socket
var currentBoard

$(document).ready(function() {
    // Connect to websocket and define message actions
    function socketConnect() {
        Socket = new WebSocket("ws://localhost:3000/ws");

        Socket.onopen = function (event) {
            if (!currentBoard) {
                boardId = $("li.menuitem")[0].id
                currentBoard = boardId
                showBoard(boardId)
            }
        }

        Socket.onmessage = function(event) {
            message = JSON.parse(event.data)
            switch(message.type) {
                case 'getBoardItems':
                    renderBoard(message.body)
                    break
                case 'newBoard':
                    addMenuItem(message.body.name, message.body.board_id)
            }
        }

        Socket.onclose = function(event) {
            setTimeout(function(){socketConnect()}, 5000);
        }
    }
    socketConnect()

    // Add some onclicks to buttons
    $(".menuitem").click(function() {
        showBoard(this.id)
    })

    $(".newboard > a").click(function() {
        $(".newboard > .popup").toggle();
        $(".popup > input").focus();
    })
    $(".newboard > .popup").keyup(function(event) {
        if(event.keyCode == 13){
            name = $(".popup > input").val();
            if (name) {
                newBoard(name);
                $(".popup > input").val("");
                $(".newboard > .popup").toggle();

            } else {
                return
            }
        }
    })

    function newBoard(name) {
        Socket.send(JSON.stringify({"type": "newBoard", "body":{"name": name}}))
    }

    function addMenuItem(name, id) {
        newItem = $("<li/>")
            .addClass("menuitem")
            .attr("id", id);
        $("<a/>").text(name)
            .appendTo(newItem)
        newItem.insertBefore($(".newboard"))
    }

    // This sends a ws request which returns 'getBoardItems'
    // in the onmessage getBoardItems causes the screen to be re-rendered
    function showBoard(boardId) {
        message = {
            "type": "getBoardItems",
            "body": {"boardId": boardId}
        }
        Socket.send(JSON.stringify(message))
        currentBoard = boardId
    }

    function mutateItem(element, command) {
        name = $(element).parent("div.row").attr("name")
        message = {
            "type": 'mutateItem',
            "body": {
                "boardId": currentBoard,
                "itemName": name,
                "delta": command
            }
        }
        if (command === "incr") {
            elem = $(element).parent("div.row");
            elem.data("proto")
                .clone(withDataAndEvents=true)
                .insertBefore($(element))
            Socket.send(JSON.stringify(message))
        } else {
            next = $(element).next()
            if (next.hasClass("box")) {
                next.remove()
                Socket.send(JSON.stringify(message))
            }
        }
    }

    renderBoard = function(boardItems) {
        main = $("div.main")
        main.empty()
        for (i = 0; i < boardItems.length; i++) {
            item = boardItems[i]
            newRow = $('<div/>')
                .addClass("row")
                .attr('name', item.name)
                .appendTo(main);

            proto = $('<div/>')
                .addClass('box')
                .addClass('c' + (i % 5))
                .click(function() {
                    mutateItem(this, 'incr')
                });
            newRow.data("proto", proto)
            $('<div/>').addClass("title")
                .text(item.name)
                .appendTo(newRow);
            newMinus = $('<div/>')
                .addClass("button")
                .appendTo(newRow);
            $("<a/>").addClass("minus")
                .appendTo(newMinus)
            newMinus.click(function() {
                mutateItem(this, 'decr')
            })
            for (j = 0; j < item.value; j++) {
                newRow.data("proto")
                    .clone(withDataAndEvents=true)
                    .appendTo(newRow)
            }
            newPlus = $('<div/>')
                .addClass("button")
                .appendTo(newRow);
            $('<a/>').addClass("plus")
                .appendTo(newPlus);
            newPlus.click(function() {
                mutateItem(this, 'incr')
            })
        }
    }
}
)
