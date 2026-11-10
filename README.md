# ttype

A tiny typing speed test that runs right in your terminal

## What it does

You get a screen full of random words. Start typing and the timer kicks off. Letters turn green as you get them right, red when you mess up, and the current character is highlighted so you know where you are. When the 60 seconds run out, you get your WPM and accuracy, plus a breakdown of how many keys you hit and how many were wrong.

The word list is about a thousand common English words, so you won't be fighting weird vocabulary — just your own fingers.

## Running it

You'll need Go installed (1.26+).

```bash
git clone https://github.com/nabsk911/ttype
cd ttype
go run .
```

Or build a binary you can keep around:

```bash
go build -o ttype .
./ttype
```

## Keys

| Key                | What it does                            |
| ------------------ | --------------------------------------- |
| just type          | starts the test on your first keystroke |
| `space`            | move to the next word                   |
| `backspace`        | fix the current word                    |
| `tab` then `enter` | restart with fresh words                |
| `esc` or `ctrl+c`  | quit                                    |

Note the restart combo — you have to tap `tab` first, then `enter`. That's on purpose so a stray enter doesn't wipe your run.

## Built with

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) for the TUI
- [Bubbles](https://github.com/charmbracelet/bubbles) for the timer component
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) for all the colors and the rounded results box
