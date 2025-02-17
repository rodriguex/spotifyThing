# SpotifyThing

This is an app for lazy people. People that like things to be fast, light and want shit to just work. I made this because I was tired of AltTabing into Spotify's client whenever I wanted to change a song. Now I just need to press a keybind, type a song and boom! I'm listening.

## How to use it

Right now probably just works on Linux, I've tested only in there, but if you have a Mac you can try to make it work. Windows sucks, so whatever.

This app uses Fyne for the UI part, so you need to install some dependencies it needs.

Install these packages for Ubuntu/Debian like systems: 

`sudo apt-get install golang gcc libgl1-mesa-dev xorg-dev libxkbcommon-dev`

If you're running other kind of Linuxes, go to these page and download the necessary packages for your system: 

https://docs.fyne.io/started/

After this, you'll need to create a new application in Spotify's Developers Portal. Sign in and go to your Dashboard:

https://developer.spotify.com/dashboard

Make a new app with the name and description you want, set the redirect_uri for http://localhost:8080/authorize and check the 'Web Api'. After created, go to settings and copy the following data into an .env file you need to create in the project:

CLIENT_ID=YOUR_CLIENT_ID  
CLIENT_SECRET=YOUR_CLIENT_SECRET  
REDIRECT_URL=YOUR_REDIRECT_URL   
SERVER_URL=YOUR_SERVER_URL   

If your redirect_url is: http://localhost:8080/authorize, server url needs to be localhost:8080.

After that, run `go build -x` (-x) to see the logs so it doenst seem that slow the process. When it's done, ./spotifyThing and search a song to see
the shit working.
