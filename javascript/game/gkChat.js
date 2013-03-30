/*
    Copyright 2012-2013 1620469 Ontario Limited.

    This program is free software: you can redistribute it and/or modify
    it under the terms of the GNU Affero General Public License as published by
    the Free Software Foundation, either version 3 of the License, or
    (at your option) any later version.

    This program is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU Affero General Public License for more details.

    You should have received a copy of the GNU Affero General Public License
    along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

var gkChatContext = new gkChatContextDef();

function gkChatContextDef() {
	this.chatLines = 24;
}

function gkChatInit() {
	var i;
	var chatDiv = document.getElementById("chatDiv");

	for (i = 0;i < gkChatContext.chatLines;i++) {
		var div1;
		var div2;

		div1 = document.createElement('div');
		chatDiv.appendChild(div1);

		div2 = document.createElement('div');
		div2.setAttribute('class','chatTime');
		div2.setAttribute('id','chatTime_' + i);
		div1.appendChild(div2);

		div2 = document.createElement('div');
		div2.setAttribute('class','chatUserBlank');
		div2.setAttribute('id','chatUser_' + i);
		div1.appendChild(div2);

		div2 = document.createElement('div');
		div2.setAttribute('class','chatMessage');
		div2.setAttribute('id','chatMessage_' + i);
		div1.appendChild(div2);
	}	
}

function gkChatSubmit() {
	var inputText = document.getElementById("chatInput");
	var message = inputText.value.replace("~",".");
	if (message.search("<") + message.search(">") != -2) {
		if (confirm("Your message may contain HTML tags. These tags, if any, will be shown in their plaintext form. If you want to submit your message as-is, press OK. Press Cancel to edit your message.") == false) {
			return false;
		}
		message = message.replace("<", "&lt;");
		message = message.replace(">", "&gt;");
	}

	var jsonMessage = JSON.stringify({ userName: gkWsContext.userName, message: message });
	if (message.length > 0) {
		gkWsSendMessage("chatReq~" + jsonMessage + "~");
	}

	inputText.value = "";
	return false;
}

function gkChatMessageFromServer(userName, message) {
    var i
    var timeSpan1
    var timeSpan2
    var userSpan1
    var userSpan2
    var messageSpan1
    var messageSpan2

    for (i = (gkChatContext.chatLines - 2);i > 0;i--) {
        timeSpan1 = document.getElementById("chatTime_" + i);
        userSpan1 = document.getElementById("chatUser_" + i);
        messageSpan1 = document.getElementById("chatMessage_" + i);
        timeSpan2 = document.getElementById("chatTime_" + (i + 1));
        userSpan2 = document.getElementById("chatUser_" + (i + 1));
        messageSpan2 = document.getElementById("chatMessage_" + (i + 1));

        timeSpan2.innerHTML = timeSpan1.innerHTML;
        userSpan2.innerHTML = userSpan1.innerHTML;
        messageSpan2.innerHTML = messageSpan1.innerHTML;

		if (userSpan2.innerHTML.length > 0) {
			userSpan2.setAttribute('class','chatUser');
		}
    }

    var d = new Date();
    timeSpan1 = document.getElementById("chatTime_1");
    timeSpan1.innerHTML = d.toLocaleTimeString();
    userSpan1 = document.getElementById("chatUser_1");
    userSpan1.innerHTML = userName
	userSpan1.setAttribute('class','chatUser');
    messageSpan1 = document.getElementById("chatMessage_1");
    messageSpan1.innerHTML = message
}

