<img src="_img/kaze.png" width="150px" align="left"/>

# kaze

Lightweight notification daemon for Linux.
Implements the [freedesktop notification spec](https://developer.gnome.org/notification-spec/).
In a usable state, but lacks features like icon & sound support, and proper markup handling.
These might be implemented in the future.

## screenshot

<img src="_img/scrot.png"/>

## features

 - config file
 - (mouse) `1`: context menu - use external program like dmenu to select link/action
 - (mouse) `3`: close notification
 - (kbd) `Mod3-Shift-Space`: close topmost notification
 - (kbd) `Mod3-Space`: show all notifications
 - notifications shown in order of importance then recency
 - fallback font support
 - can limit memory usage

## todo

 - move config file to sensible location
 - maybe use fontconfig so only need to give font name

