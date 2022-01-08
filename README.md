# Palette
## A Multi-Purpose WebApp Built Around Drawing
![Home Page](https://github.com/albertbregonia/Palette/blob/main/img/home.png?raw=true "Home Page")

**Palette** is a small yet powerful WebApp built primariy in Golang. It is designed to create a web server in which people can easily join and create lobbies to draw. Currently, the system features a streamable whiteboard where the host of a lobby can free draw for everyone in the lobby and a customizable pictionary mode. The customizable whiteboard makes use of many features within JavaScript in order to allow the artist to truly lay their mind on the page.

![Blackboard](https://github.com/albertbregonia/Palette/blob/main/img/blackboard.png?raw=true "Blackboard")

![Whiteboard](https://github.com/albertbregonia/Palette/blob/main/img/whiteboard.png?raw=true "Whiteboard")

After learning about the HTML canvas and the capabilities that it held paired with JavaScript, I initially intended on making a [skribbl.io](https://skribbl.io/ "Skribbl.io by @ticedev on Twitter") clone. After spending countless hours for one week straight, coding and learning about JavaScript/Golang, I decided to make this WebApp much more. As of 3/1/21, this is my most advanced project.

***Important Note: I am currently creating a v2 of this repository that will provide a better system in terms of custom game creation and utilize WebRTC DataChannels for better performance***

# Main Features
- Customizable Whiteboard with Dark/Light Theme
- Touch screen support
- If players leave during the game, they have a 5 minute time window to return before their data is deleted

# Future Features
- Game Template to easily allow development of new game modes
- More Included Game Modes
- Flood Fill Algorithm (Fill Bucket Tool)

# Dependencies and Requirements
- Designed using Golang 1.17
- HTML5, CSS3, JavaScript (ES8); Essentially, a modern browser is required to use the `<canvas>` in HTML
- [gorilla/sessions](https://github.com/gorilla/sessions "Sessions by The Gorilla Team")
- [gorilla/websockets](https://github.com/gorilla/websocket "WebSocket by The Gorilla Team")
- [pion/webrtc](https://github.com/pion/webrtc "WebRTC by Pion")
