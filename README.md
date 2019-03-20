# kaze - 風

Tiny and lightweight notification service for Linux.
Implements the [freedesktop notification spec](https://developer.gnome.org/notification-spec/#hints).
Massive WIP.
Idea: get notification, display a popup, then put an indicator in your bar.
Then you hit a keybinding to view/close notifications.

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
 - [ ] support non-ascii
 - [x] config file
 - [ ] _maybe:_
   - [ ] sound support
   - [ ] icon support
   - [ ] better font support (currently you need to pass the full path to the ttf),
   maybe need to use fontconfig to find the correct font
