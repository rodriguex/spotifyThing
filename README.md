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

Go to 'Create App'. You can name it how you want, what really matters in this part is that the REDIRECT_URI you put matches with the one in the code. By default, I set it for localhost:8080/authorize. If you want to change this, don't forget to change also in the code. `main.go`, line 28 in REDIRECT_URI global variable.