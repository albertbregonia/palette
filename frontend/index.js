//temporary! these functions are merely implemented to test the backend

const lobbyNameInput = document.getElementById(`lobby-name`),
      usernameInput = document.getElementById(`username`),
      passwordInput = document.getElementById(`password`),
      whiteboard = document.getElementById(`whiteboard`),
      whiteboardStream = document.getElementById(`whiteboard-stream`);

(function reconnectHandler() {
    fetch(`/reconnect`, {method: `post`})
    .then(response => {
        if(response.status == 202) {
            WebRTCStartup();
            return response.text();
        }
    }).then(msg => msg ? alert(msg) : undefined).catch(alert);
})();

function loginHandler(createLobby) {
    if(![lobbyNameInput.value, usernameInput.value, passwordInput.value].every(e => e))
        return false;
    const info = new URLSearchParams({
        lobby: lobbyNameInput.value,
        password: passwordInput.value,
        username: usernameInput.value,
        create: !!createLobby
    });
    fetch(`/login?${info}`, {method: `post`})
    .then(response => {
        if(response.status == 202)
            WebRTCStartup();
        return response.status;
    })
    .then(alert)
    .catch(alert);
    return false;
}

function disconnectHandler() {
    fetch(`leave`)
    .then(response => alert(response.status) || location.reload())
    .catch(alert);
}

const chatInput = document.getElementById(`chat-input`);
function chat() {
    rtc.chat.send(JSON.stringify({
        content: chatInput.value,
    }));
    chatInput.value = ``;
    return false;
}

function whiteboardSetup() {
    whiteboard.brush = whiteboard.getContext(`2d`);
    whiteboard.brush.fillStyle = `white`;
    whiteboard.brush.fillRect(0, 0, whiteboard.width, whiteboard.height);
    whiteboard.brush.lineWidth = 5;
    whiteboard.brush.lineCap = `round`;
    whiteboard.addEventListener(`mousedown`, startDrawing);
    whiteboard.addEventListener(`mouseup`, stopDrawing);
    whiteboard.addEventListener(`mouseleave`, stopDrawing);
    whiteboard.addEventListener(`mousemove`, drawHandler);
    whiteboard.addEventListener(`touchmove`, drawHandler);
    whiteboard.addEventListener(`touchstart`, startDrawing);
    whiteboard.addEventListener(`touchend`, stopDrawing);
    whiteboard.addEventListener(`touchcancel`, stopDrawing);
}

function startDrawing(e) {
    whiteboard.isDrawing = true;
    drawHandler(e);
}

function stopDrawing() {
    whiteboard.isDrawing = false;
    whiteboard.brush.beginPath();
}

function drawHandler(e) {
    e.preventDefault();
    if(whiteboard.isDrawing) {
        let x = e.clientX - whiteboard.offsetLeft;
        let y = e.clientY - whiteboard.offsetTop;
        let clientX = e.clientX, clientY = e.clientY;
        if (e.touches && e.touches.length == 1) {
            let touch = e.touches[0];
            x = touch.pageX - whiteboard.offsetLeft;
            y = touch.pageY - whiteboard.offsetTop;
            clientX = touch.pageX;
            clientY = touch.pageY;
        }
        whiteboard.brush.lineTo(x,y);
        whiteboard.brush.stroke();
        whiteboard.brush.beginPath();
        whiteboard.brush.moveTo(x,y);
    }
}