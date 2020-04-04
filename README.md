<div align="center">
    <img src="assets/logo.svg" />
    <p>An <b>AUR</b> helper and more</p>
</div>

[![AUR version](https://img.shields.io/aur/version/yup.svg)](https://aur.archlinux.org/packages/yup/)
[![AUR bin version](https://img.shields.io/aur/version/yup-bin?color=%230084ff&label=bin)](https://aur.archlinux.org/packages/yup-bin/)
[![GitHub](https://img.shields.io/github/license/ericm/yup.svg)](https://github.com/ericm/yup/blob/master/LICENSE)
[![GitHub contributors](https://img.shields.io/github/contributors/ericm/yup.svg)](https://github.com/ericm/yup/graphs/contributors)

**Yup** helps you install packages with ease on Arch Linux

## Features

- Searching with `yup [search-terms]` returns most accurate results
  ![](assets/scr1.png?raw=true)

- Uses _ncurses_ to display search results. This allows for mouse interaction in the terminal and easier navigation.
  [![asciicast](https://asciinema.org/a/VGzR3JYAjGqT91SfiKjBUfFkh.svg)](https://asciinema.org/a/VGzR3JYAjGqT91SfiKjBUfFkh)

- Don't want to use ncurses? Use `yup -n` to use non-ncurses mode

- Want to search the AUR exclusively? Use `yup -a`

- Like _yay_, type `yup` to run a system upgrade.

- An easy to use config file located at `~/.config/yup/config.json` in JSON format.

* Want to see which packages are cluttering up your system? Run `yup -Qos` to get a list ordered package size.

## Configuration

- Config file found at `~/.config/yup/config.json`.
- The config file has the following options:
  ```
{
  "sort_mode": "closest",
  "ncurses_mode": true,
  "always_update_repos": false,
  "print_pkgbuild": true,
  "ask_pkgbuild": true,
  "ask_redo": true,
  "version": "1.1.1",
  "silent_update": true,
  "pacman_limit": 200,
  "aur_limit": 200,
  "vim_keybindings": false
}
  ```

## Usage

```
    yup                 Updates AUR and pacman packages (Like -Syu)
    yup <package(s)>    Searches for that packages and provides an install dialogue
Operations:
    yup {-h --help}
    yup {-V --version}
    yup {-D --database} <options> <package(s)>
    yup {-F --files}    <options> <package(s)>
    yup {-Q --query}    <options> <package(s)>
    yup {-R --remove}   <options> <package(s)>
    yup {-S --sync}     <options> <package(s)>
    yup {-T --deptest}  <options> <package(s)>
    yup {-U --upgrade}  <options> <file(s)>
Custom operations:
    yup -c              Cleans cache and unused dependencies
    yup -C              Cleans AUR cache only
    yup -a [package(s)] Operates on the AUR exclusively
    yup -n [package(s)] Runs in non-ncurses mode
    yup -Y <Yupfile>    Install packages from a Yupfile
    yup -Qos            Orders installed packages by install size
```

## Differences between yay or trizen

- Yup gives you the **most accurate results** first. As seen in the example above, yup sorts the results to bring the most accurate to the start.

- `Yupfiles` are small files that allow you to batch install packages with a single command. [Here's an example Yupfile](test.Yupfile)

- Yup uses _ncurses_. This allows users to both scroll while not displacing the bottom bar and easily navigate to certain results using more natural forms of user input.

- Yup has an easy config file seperate to that of pacman's. This allows it to be more customisable from the get go.

- Yup has both `yup -c` (for clearing all package cache) and yup `yup -C` (for clearing yup's cache only).

- Yup allows you to disable ncurses mode (to normal terminal output) using `yup -n` temporarily or permanently by changing a value in the config file.

- In the search menu, yup allows you to remove an installed package instantly using the `R` hotkey.

- After selecting packages to install, you can revise your decision if you made a mistake.

- Yup will _soon_ allow you to disable any of the dialogue during install using the config menu.

## Installing

### From the AUR

1. `git clone https://aur.archlinux.org/yup.git`
2. `cd yup`
3. `makepkg -si`

### From the AUR (binary)

1. `git clone https://aur.archlinux.org/yup-bin.git`
2. `cd yup-bin`
3. `makepkg -si`

### From source

Make sure you have `go>=1.12`, `ncurses` and `make`.

1. Clone the repo
2. Run `make`
3. Install with `make install`

### Completions not working on zsh

- You'll need to add `compaudit && compinit` to the bottom of your .zshrc

## Credits

Copyright 2019 Eric Moynihan

Inspired by [Jguer](https://github.com/Jguer)'s [yay](https://github.com/Jguer/yay)
