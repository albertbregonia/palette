# Palette
## A Multi-Purpose WebApp Built Around Drawing
![Home Page](https://github.com/albertbregonia/Palette/blob/main/img/home.png?raw=true "Home Page")

**Palette** is a small yet powerful WebApp built primariy in Golang. It is designed to create a web server in which people can easily join and create lobbies to draw. Currently, the system features a streamable whiteboard where the host of a lobby can free draw for everyone in the lobby and a customizable pictionary mode. The customizable whiteboard makes use of many features within JavaScript in order to allow the artist to truly lay their mind on the page.

![Blackboard](https://github.com/albertbregonia/Palette/blob/main/img/blackboard.png?raw=true "Blackboard")

![Whiteboard](https://github.com/albertbregonia/Palette/blob/main/img/whiteboard.png?raw=true "Whiteboard")

After learning about the HTML canvas and the capabilities that it held paired with JavaScript, I initially intended on making a [skribbl.io](https://skribbl.io/ "Skribbl.io by @ticedev on Twitter") clone. After spending countless hours for one week straight, coding and learning about JavaScript/Golang, I decided to make this WebApp much more. As of 3/1/21, this is my most advanced project.

# Main Features:
- Customizable Whiteboard
  - Brush Size Slider - Range of `1px` to `100px` to allow for more precision
  - Color - All Standard Colors in HTML are readily available
- Touch screen support
- Dark/Light Theme
- Easy of Use
  - Lobby Generation and Deletion - any user can create/join a lobby given a name and password
  - If players leave during the game, they have a 5 minute time window to return before their data is deleted

# Future Features:
- Game Template to easily allow development of new game modes
- More Included Game Modes
- Shareable Lobby Links
- Embedded Files into Final Executable - Golang 1.16
- Flood Fill Algorithm (Fill Bucket Tool)

# Installation and Requirements:
- Designed using Golang 1.15.6 *Current Build is 1.15.6*
- HTML5, CSS3, JavaScript; Essentially, a modern browser is required to use the `<canvas>` in HTML
- [Gorilla Sessions](https://github.com/gorilla/sessions "Sessions by The Gorilla Team")
- [Gorilla WebSocket](https://github.com/gorilla/websocket "WebSocket by The Gorilla Team")

# Known Limitations:
- ***Like with any other WebApp, 9 times out of 10, a simple refresh will fix your issue. :)***
- **Cursor Offset**: Upon resizing the window, an offset between the cursor and the paintbrush is created. Correcting this offset in JavaScript clears the canvas data therefore, I have decided to refrain from correcting it.
- **Visibility**: I wanted to allow maximum screen real estate for users to draw. However, this does not work well with other uses as the canvas data will not scale in HTML. There are two remedies for this. One remedy is to start drawing in the top left of the canvas. This is a region that all users, **no matter what resolution or window scaling**, will be able to see. Another remedy is to simply coordinate with your group and let them know which regions are/are not visible on all of their displays and to draw that border.
- **Performance**: Although this uses Golang and performance is not an issue, I have not benchmarked this system on a larger scale. If I were to give an estimate on limits, I would say that a lobby could handle around 10 concurrent lobbies with 20 players each.
