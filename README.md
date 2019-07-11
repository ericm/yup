<div align="center">
    <img src="assets/logo.svg" />
    <p>An <b>AUR</b> helper and more</p>
</div>

**Yup** helps you install packages with ease on Arch Linux

## Features
- Searching with `yup [search-terms]` returns most accurate results
![](assets/scr1.png?raw=true)

- Uses *ncurses* to display search results. This allows for mouse interaction in the terminal and easier navigation.
![](assets/scr2.gif?raw=true)
- Don't want to use ncurses? Use `yup -n` to use non-ncurses mode

- Want to search the AUR exclusively? Use `yup -a`

- Like *yay*, type `yup` to run a system upgrade.

- An easy to use config file located at `~/.config/yup/yup.conf` in JSON format.

- Want to see which packages are cluttering up your system? Run `yup -Qos` to get a list ordered package size.

## Installing
### From the AUR
1. `git clone https://aur.archlinux.org/yup.git`
2. `cd yup`
3. `makepkg -si`

### From source
Make sure you have `go>=1.12`, `ncurses` and `make`.
1. Clone the repo
2. Run `make`
3. Install with `make install`

## Credits
Copyright 2019 Eric Moynihan

Inspired by [Jguer](https://github.com/Jguer)'s [yay](https://github.com/Jguer/yay)
