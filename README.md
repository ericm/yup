<div align="center">
    <img src="assets/logo.svg" />
    <p>An <b>AUR</b> helper and more</p>
</div>

[![AUR version](https://img.shields.io/aur/version/yup.svg)](https://aur.archlinux.org/packages/yup/)
![GitHub](https://img.shields.io/github/license/ericm/yup.svg)
![GitHub contributors](https://img.shields.io/github/contributors/ericm/yup.svg)

**Yup** helps you install packages with ease on Arch Linux

## Features
- Searching with `yup [search-terms]` returns most accurate results
![](assets/scr1.png?raw=true)

- Uses *ncurses* to display search results. This allows for mouse interaction in the terminal and easier navigation.
![](assets/scr2.gif?raw=true)
- Don't want to use ncurses? Use `yup -n` to use non-ncurses mode

- Want to search the AUR exclusively? Use `yup -a`

- Like *yay*, type `yup` to run a system upgrade.

- An easy to use config file located at `~/.config/yup/config.json` in JSON format.

- Want to see which packages are cluttering up your system? Run `yup -Qos` to get a list ordered package size.

## Differences between yay or trizen
- Yup gives you the **most accurate results** first. As seen in the example above, yup sorts the results to bring the most accurate to the start.

- Yup uses *ncurses*. This allows users to both scroll while not displacing the bottom bar and easily navigate to certain results using more natural forms of user input.

- Yup has an easy config file seperate to that of pacman's. This allows it to be more customisable from the get go.

- Yup has both `yup -c` (for clearing all package cache) and yup `yup -C` (for clearing yup's cache only).

- Yup allows you to disable ncurses mode (to normal terminal output) using `yup -n` temporarily or permanently by changing a value in the config file.

- In the search menu, yup allows you to remove an installed package instantly using the `R` hotkey.

- After selecting packages to install, you can revise your decision if you made a mistake.

- Yup will *soon* allow you to disable any of the dialogue during install using the config menu.

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
