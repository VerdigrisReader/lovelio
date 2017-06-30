var Socket
var newBoard = function(){};
var currentBoard

$(document).ready(function() {
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
            }
        }

        Socket.onclose = function(event) {
            setTimeout(function(){socketConnect()}, 5000);
        }
    }
    socketConnect()

    newBoard = function() {
        Socket.send(JSON.stringify({"type": "newBoard"}))
    }

    // This sends a ws request which returns 'getBoardItems'
    // in the onmessage getBoardItems causes the screen to be re-rendered
    showBoard = function(boardId) {
        message = {
            "type": "getBoardItems",
            "body": {"boardId": boardId}
        }
        Socket.send(JSON.stringify(message))
        currentBoard = boardId
    }

    mutateItem = function(element, command) {
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
