const whiteboard = document.getElementById(`whiteboard`),
      brush = whiteboard.getContext(`2d`);
      sizer = document.getElementById(`sizer`),
      game = document.getElementById(`game`),
      login = document.getElementById(`login`),
      announcements = document.getElementById(`announcements`);

let chatWebsocket, drawingWebsocket, last, //chat & drawing server websocket, last = most recent drawing data
    isPainting = false,
    theme = `white`, //whiteboard theme
    border = `black`;

((noLogin, debugStyle) => {
    setTimeout(() => {
        game.style.display = `flex`;
    login.style.display = `none`;
    announcements.style.display = `block`;
    }, 5000);

    setTimeout(() => {
        login.style.animation = `slide-up 500ms forwards`;
    }, 4500);
    whiteboardStartup();
    addArtistEvents();
})();

//======= INTERNAL CALLS =======//

async function hasGame() { //call the backend and check for valid game, resume if true
    try {
        const response = await fetch(`game`, { method: `POST` });
        if(response.status == 202) {
            game.style.display = `flex`;
            login.style.display = `none`;
            announcements.style.display = `block`;
        } else {
            document.body.style.backgroundImage = `url(https://images.pexels.com/photos/346529/pexels-photo-346529.jpeg?auto=compress&cs=tinysrgb&dpr=2&h=750&w=1260)`;
            document.body.style.backgroundSize = `cover`;
            document.body.style.backgroundRepeat = `no-repeat`;
        }
    } catch(e) { console.log(e); }
}

function lobbyRequest(url) {
    let info = document.getElementById(`lobbyInfo`);
    let end = false;
    Array.from(info.getElementsByTagName(`input`)).forEach(input => {
        if(input.value.trim()==`` && !end) { //check for valid info given
            console.log(`Invalid Value for: ${input.placeholder}`);
            end = true;
        }
    })
    if(!end) //if parameters are not empty
        fetch(url, { //attempt a lobby request
            method: `POST`,
            body: new URLSearchParams(new FormData(info))
        }).then(r => {
            switch(r.status) {
                case 202: location.replace(`/`);                           break;
                case 401: console.log(`Incorrect password for existing lobby.`); break;
                case 404: console.log(`Lobby Not Found`);                        break;
                default: console.log(r.statusText);
            }
        }).catch(e => console.log(e));
    return false;
}

function leaveLobby() {
    chatWebsocket.send(String.fromCharCode(8)); //send deliberate disconnect request
    fetch(`leave`, { //clear cookies for server database
        method: `POST`
    }).then(() => location.reload())
}

function addArtistEvents() { //enable drawing
    whiteboard.addEventListener(`mousedown`, start);
    whiteboard.addEventListener(`mouseup`, stop);
    whiteboard.addEventListener(`mouseleave`, stop);
    whiteboard.addEventListener(`mousemove`, draw);
    whiteboard.addEventListener(`touchmove`, draw);
    whiteboard.addEventListener(`touchstart`, start);
    whiteboard.addEventListener(`touchend`, stop);
    whiteboard.addEventListener(`touchcancel`, stop);
    whiteboard.addEventListener(`dblclick`, themeSwitch);
    document.getElementById(`clear`).style.display = `block`;
    document.getElementById(`toolbar`).style.display = `inline-block`;
}

function removeArtistEvents() { //disable drawing
    whiteboard.removeEventListener(`mousedown`, start);
    whiteboard.removeEventListener(`mouseup`, stop);
    whiteboard.removeEventListener(`mouseleave`, stop);
    whiteboard.removeEventListener(`mousemove`, draw);
    whiteboard.removeEventListener(`touchmove`, draw);
    whiteboard.removeEventListener(`touchstart`, start);
    whiteboard.removeEventListener(`touchend`, stop);
    whiteboard.removeEventListener(`touchcancel`, stop);
    whiteboard.removeEventListener(`dblclick`, themeSwitch);
    document.getElementById(`clear`).style.display = `none`;
    document.getElementById(`toolbar`).style.display = `none`;
}

function connectionLost() {
    console.log(`You have lost connection to the lobby due to an unfortunate error.`);
    location.reload();
}

//create a websocket to connect to the lobby's chat thread; parse responses
function chatStartup() {
    chatWebsocket = new WebSocket(`ws://${location.hostname}:${location.port}/chat`);
    chatWebsocket.onopen = () => { console.log(`Chat successfully initialized`); drawServerStartup();}
    chatWebsocket.onerror = connectionLost;
    chatWebsocket.onmessage = msg => {
        let text = msg.data;
        switch(text.substring(0,6)) {
            case `<anmt>`:
                if (text.indexOf(`Round Over.`) > -1 || text.indexOf(`Game Starting`)) {
                    stop();
                    wipeBoard();
                    removeArtistEvents();
                    document.getElementById(`clear`).style.display = `none`;
                }
                let header = document.getElementById(`announcements`);
                header.innerHTML = text.substring(6, text.length);
                break;
            case `<draw>`:
                addArtistEvents();
                toChatMsg(msg.data);
                break;
            default:
                toChatMsg(msg.data);
        }
    };
}

//======= DOM MANIPULATION =======//

function toggleMoreColors(e) {
    let more = document.getElementById(`moreColors`); 
    more.style.display = (more.style.display == `flex`) ? `none` : `flex`;
}

function themeSwitch() {
    switch(theme) {
        case `black`:
            theme = `white`;
            border = `black`;
            break;
        case `white`:
            theme = `black`;
            border = `white`;
            break;
    }
    Array.from(document.getElementsByTagName(`box`)).forEach(box => box.style.border = `1px solid ${border}`);
    last = {'id': 3}
    drawingWebsocket.send(JSON.stringify(last))
    updateCursor();
    wipeBoard();
}

function chatMessage() { //send new message to lobby
    let input = document.getElementById(`msg`);
    if (input.value.trim() != ``) {
        chatWebsocket.send(input.value.trim());
        input.value = ''; //clear previous text
    }
    return false;
}

function toChatMsg(s) { //string to chat msg in html
    let log = document.getElementById(`log`);
    let msg = document.createElement(`li`);
    msg.innerHTML = s;
    log.appendChild(msg);
    log.scrollTo(0, log.scrollHeight); //autoscroll to new message
}

//======= DRAWING =======//

//set main properties of the whiteboard and connect to chat
function whiteboardStartup() {
    console.log(`Initializing Whiteboard...`);
    whiteboard.height = 3840;
    whiteboard.width = 2160;
    brush.lineWidth = sizer.value;
    brush.lineCap = `round`;
    wipeBoard()
    sizer.addEventListener(`change`, updateCursor);
    Array.from(document.getElementsByTagName(`box`)).forEach(color => {
        color.addEventListener(`click`, () => {
            brush.strokeStyle = color.style.backgroundColor;
            document.getElementById(`expandColor`).style.backgroundColor = color.style.backgroundColor;
            updateCursor();
            Array.from(document.getElementById(`toolbar`).getElementsByTagName(`box`)).forEach(box => {
                box.classList.remove(`selected`);
            });
            color.classList.add(`selected`);
        });
    });
    chatStartup();
}

//uses a hidden canvas to generate the cursor given the current brush settings
function updateCursor() {
    let cursor = document.getElementById(`cursor`), ctx = cursor.getContext('2d');
    let r = sizer.value/2;
    ctx.clearRect(0, 0, cursor.width, cursor.height); //clear the old cursor
    ctx.beginPath();
    ctx.arc(50, 50, r, 0, 2 * Math.PI, false); //draw a single circle of the newly selected size
    ctx.fillStyle = brush.strokeStyle;
    ctx.fill();
    ctx.lineWidth  = brush.lineWidth = sizer.value; //update settings
    ctx.lineWidth /= 16; //draw accent border for visibility
    ctx.strokeStyle = border;
    ctx.stroke();
    whiteboard.style.cursor = `url('${cursor.toDataURL()}') 50 50, auto`; //change cursor and center
}

//start drawing
function start(e) {
    isPainting = true;
    draw(e);
}

//end brush stroke
function stop() { 
    isPainting = false;
    brush.beginPath();
    last = {'id': 1};
    drawingWebsocket.send(JSON.stringify(last));
}

//take in drawing data from either the mouse/touch screen and broadcast it to the lobby; essentially tracks continuous movement
function draw(e) {
    e.preventDefault();
    if(isPainting) {
        let x = e.clientX-whiteboard.offsetLeft;
        let y = e.clientY-whiteboard.offsetTop;
        let clientX = e.clientX, clientY = e.clientY;
        if (e.touches && e.touches.length == 1) {
            let touch = e.touches[0];
            x = touch.pageX - whiteboard.offsetLeft;
            y = touch.pageY - whiteboard.offsetTop;
            clientX = touch.pageX;
            clientY = touch.pageY;
        }
        brush.lineTo(x,y);
        brush.stroke();
        brush.beginPath();
        brush.moveTo(x,y);
        last = {
            'id': 0,
            'clientX': clientX,
            'clientY': clientY,
            'originalX': whiteboard.offsetLeft,
            'originalY': whiteboard.offsetTop,
            'width': sizer.value,
            'color': brush.strokeStyle
        };
        drawingWebsocket.send(JSON.stringify(last));
    }
}

//clear the whiteboard given the current theme
function wipeBoard() {
    brush.fillStyle = theme;
    brush.fillRect(0, 0, whiteboard.width, whiteboard.height);
    last = {'id': 2};
    if (drawingWebsocket!=null)
        drawingWebsocket.send(JSON.stringify(last));
}

//create secondary websocket after chat to read in echoed JSON data; simulates drawing
function drawServerStartup() {
    drawingWebsocket = new WebSocket(`ws://${location.hostname}:${location.port}/draw`);
    drawingWebsocket.onopen = () => console.log(`Draw Server successfully initialized`);
    drawingWebsocket.onerror = connectionLost;
    drawingWebsocket.onmessage = msg => {
        let data = JSON.parse(msg.data);
        if (data != last) {
            switch(parseInt(data.id)) {
                case 0: //simulate drawing data
                    let x = data[`clientX`]-data[`originalX`];
                    let y = data[`clientY`]-data[`originalY`];
                    brush.strokeStyle = data[`color`];
                    brush.lineWidth = data[`width`];
                    brush.lineTo(x,y);
                    brush.stroke();
                    brush.beginPath();
                    brush.moveTo(x,y);
                    break;
                case 1: //stop drawing
                    brush.beginPath();
                    break;
                case 2: //clear canvas
                    brush.fillStyle = theme;
                    brush.fillRect(0, 0, whiteboard.width, whiteboard.height);
                    break;
                case 3: //change theme
                    switch(theme) {
                        case `black`:
                            theme = `white`;
                            border = `black`;
                            break;
                        case `white`:
                            theme = `black`;
                            border = `white`;
                            break;
                    }
                    Array.from(document.getElementsByTagName(`box`)).forEach(box => box.style.border = `1px solid ${border}`);
                    updateCursor();
                    break;
            }
        }
    };
}