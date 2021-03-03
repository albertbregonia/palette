# Palette
## A Multi-Purpose WebApp Built Around Drawing
**Palette** is a small yet powerful WebApp built primariy in Golang. It is designed to create a web server in which people can easily join and create lobbies to draw. Currently, the system features a streamable whiteboard where the host of a lobby can free draw for everyone in the lobby and a customizable pictionary mode. The customizable whiteboard makes use of many features within JavaScript in order to allow the artist to truly lay their mind on the page.

After learning about the HTML canvas and the capabilities that it held paired with JavaScript, I initially intended on making a [skribbl.io](https://skribbl.io/ "Skribbl.io by @ticedev on Twitter") clone. After spending countless hours for one week straight, coding and learning about JavaScript/Golang, I decided to make this WebApp much more. As of 3/1/21, this is my most advanced project.

# Main Features:

# Installation and Requirements:

# Known Limitations
- ***Like with any other WebApp, 9 times out of 10, a simple refresh will fix your issue. :)***
- **Cursor Offset**: Upon resizing the window, an offset between the cursor and the paintbrush is created. Correcting this offset in JavaScript clears the canvas data therefore, I have decided to refrain from correcting it.
- **Visibility**: I wanted to allow maximum screen real estate for users to draw. However, this does not work well with other uses as the canvas data will not scale in HTML. There are two remedies for this. One remedy is to start drawing in the top left of the canvas. This is a region that all users, **no matter what resolution or window scaling**, will be able to see. Another remedy is to simply coordinate with your group and let them know which regions are/are not visible on all of their displays and to draw that border.

# Dependencies:
- [Gorilla Sessions](https://github.com/gorilla/sessions "Sessions by The Gorilla Team")
- [Gorilla WebSocket](https://github.com/gorilla/websocket "WebSocket by The Gorilla Team")
