//temporary! these functions are merely implemented to test the backend

const loginDialog = document.getElementById(`login-dialog`),
      lobbyNameInput = document.getElementById(`lobby-name`),
      usernameInput = document.getElementById(`username`),
      passwordInput = document.getElementById(`password`),
      mainUI = document.getElementById(`main-ui`);

// user login and lobby registration

(function reconnectHandler() {
    fetch(`/reconnect`, {method: `post`})
    .then(response => {
        if(response.status == 202) {
            WebRTCStartup();
            (async () => alert(await response.text()))();
        }
    });
})();

function loginHandler(createLobby) {
    loginDialog.classList.add(`fade-up-out`);
    mainUI.style.display = `flex`;
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
        if(response.status == 202) {
            WebRTCStartup();
        }
    });
    return false;
}

function disconnectHandler() {
    fetch(`leave`)
    .then(response => alert(response.status) || location.reload())
    .catch(alert);
}

// set up connections

function WebRTCStartup() {

    function formatSignal(event, data) {
        return JSON.stringify({ 
            event: event, 
            data: JSON.stringify(data)
        });
    }

    const ws = new WebSocket(`wss://${location.hostname}:${location.port}/connect`); //create a websocket for WebRTC signaling 
    ws.onopen = () => console.log(`Connected`);
    ws.onclose = ws.onerror = ({reason}) => alert(`Disconnected ${reason}`);
    
    const rtc = new RTCPeerConnection({iceServers: [{urls: `stun:stun.l.google.com:19302`}]}); //create a WebRTC instance
    rtc.onicecandidate = ({candidate}) => candidate && ws.send(formatSignal(`ice`, candidate)); //if the ice candidate is not null, send it to the peer
    rtc.oniceconnectionstatechange = () => rtc.iceConnectionState == `failed` && rtc.restartIce();
    rtc.ondatachannel = ({channel}) => {
        if(channel.label != `whiteboard`)
            return;
        whiteboardSetup();
        rtc.whiteboard = channel;
        rtc.whiteboard.onmessage = ({data}) => shareHandler(JSON.parse(data));
    };

    ws.onmessage = async ({data}) => { //signal handler
        const signal = JSON.parse(data),
              content = JSON.parse(signal.data);
        switch(signal.event) {
            case `offer`:
                console.log(`got offer!`, content);
                await rtc.setRemoteDescription(content); //accept offer
                const answer = await rtc.createAnswer();
                await rtc.setLocalDescription(answer);
                ws.send(formatSignal(`answer`, answer)); //send answer
                console.log(`sent answer!`, answer);
                break;
            case `ice`:
                console.log(`got ice!`, content);
                rtc.addIceCandidate(content); //add ice candidates
                break;
            default:
                console.log(`Invalid message:`, content);
        }
    };
}

// set up drawing on the whiteboard

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
        if (e.touches && e.touches.length == 1) {
            let touch = e.touches[0];
            x = touch.pageX - whiteboard.offsetLeft;
            y = touch.pageY - whiteboard.offsetTop;
        }
        whiteboard.brush.lineTo(x, y);
        whiteboard.brush.stroke();
        whiteboard.brush.beginPath();
        whiteboard.brush.moveTo(x, y);
    }
}