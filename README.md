<div align="center">
    <img src="assets/logo.svg" />
    <p>An <b>AUR</b> helper and more</p>
</div>

**Yup** helps you install packages with ease on Arch Linux

## Features
- Searching with `yup [search-terms]` returns most accurate results
![](assets/scr1.png?raw=true)

- Uses *ncurses* to display search results. This allows for mouse interaction in the terminal and easier navigation.

- Don't want to use ncurses? Use `yup -n` to use non-ncurses mode

- Want to search the AUR exclusively? Use `yup -a`

- Want to see which packages are cluttering up your system? Run `yup -Qos` to get a list ordered package size.

## Installing
Make sure you have `go>=1.12`, `ncurses` and `make`.
1. Clone the repo
2. Run `make`
3. Install with `make install`