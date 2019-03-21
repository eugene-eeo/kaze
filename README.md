<p align="center"><img src="_img/kaze.png" width="200px"/></p>

# kaze - 風

Tiny and lightweight notification daemon for Linux.
Implements the [freedesktop notification spec](https://developer.gnome.org/notification-spec/).
Massive WIP.

## screenshot

<img src="_img/scrot.png"/>

## roadmap

 - [x] draw notification in X window
 - [x] close popups automatically
 - [x] view all notifications
 - [x] support actions
 - [x] support hyperlinks
 - [x] render notifications properly
 - [ ] support push-based client to get no. of unread notifications
 - [ ] support non-ascii (probably need to move to freetype rasterizer?)
 - [x] config file
 - [ ] _maybe:_
   - [ ] sound support
   - [ ] icon support
   - [ ] better font support (currently you need to pass the full path to the ttf),
   maybe need to use fontconfig to find the correct font
   - [ ] italic / bold / underline rendering
