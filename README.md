# 🦎 Gecko Text Editor

> **Work in Progress**: This project is currently under active development. Features may be incomplete, and breaking changes are expected. Contributions and feedback are welcome!

## Features

### Core Editing
- **Syntax Highlighting**: Powered by [Chroma](https://github.com/alecthomas/chroma) with support for numerous programming languages
- **Multi-line Selection**: Select text with Shift + Arrow keys or Ctrl + Shift + Arrow for word-by-word selection
- **Undo/Redo**: Full undo/redo history for all text operations
- **Smart Navigation**: Jump to line, find text with live results, and navigate by words or pages

### Text Operations
- **Clipboard Integration**: Copy, cut, and paste with both system and internal clipboard support
- **Find and Replace**: Interactive search with result navigation and live preview
- **Go to Line**: Quick navigation to specific line numbers
- **Select All**: Instant selection of entire document content

### User Interface
- **Status Bar**: Shows file status, cursor position, and contextual messages
- **Line Numbers**: Clear line numbering with syntax-aware highlighting
- **Help System**: Built-in help overlay with all keyboard shortcuts
- **Responsive Design**: Adapts to terminal window size changes

### File Management
- **Auto-save Detection**: Tracks file modifications with visual indicators
- **Multiple File Formats**: Supports various file types with appropriate syntax highlighting

## Installation

### Prerequisites
- Go 1.19 or higher
- `xclip` (for Linux clipboard support)

### Build from Source
```bash
git clone https://github.com/elxgy/Gecko
cd gecko
go mod tidy
go build -o Gecko
```

### Install Dependencies
On Ubuntu/Debian:
```bash
sudo apt install xclip
```

On Arch Linux:
```bash
sudo pacman -S xclip
```

## Usage

### Basic Usage
```bash
# Open a new file
./gecko <filename>

# Open an existing file
./gecko filename.txt

# Open a source code file with syntax highlighting
./gecko main.go
```

## Keyboard Shortcuts

### File Operations
| Shortcut | Action |
|----------|--------|
| `Ctrl+S` | Save file |
| `Ctrl+Q` | Quit editor |

### Editing
| Shortcut | Action |
|----------|--------|
| `Ctrl+C` | Copy selected text |
| `Ctrl+X` | Cut selected text |
| `Ctrl+V` | Paste text |
| `Ctrl+Z` | Undo |
| `Ctrl+Y` | Redo |
| `Ctrl+A` | Select all |

### Navigation
| Shortcut | Action |
|----------|--------|
| `Ctrl+F` | Find text (opens search interface) |
| `Ctrl+N` | Find next occurrence |
| `Ctrl+L` | Find previous occurrence |
| `Ctrl+G` | Go to line |
| `↑/↓` | Navigate through search results |

### Text Selection
| Shortcut | Action |
|----------|--------|
| `Shift+Arrow` | Select text in direction |
| `Ctrl+Arrow` | Move cursor by word |
| `Alt+Arrow` | Select text by word |
| `Home/End` | Move to line start/end |
| `PgUp/PgDn` | Move by page |

### Interface
| Shortcut | Action |
|----------|--------|
| `Ctrl+H` | Toggle help overlay |
| `Enter` | Confirm dialog actions |
| `Escape` | Cancel dialog/search |

## Syntax Highlighting

Gecko supports syntax highlighting for numerous programming languages including:
- Go
- Python
- JavaScript/TypeScript
- C/C++
- Rust
- Java
- HTML/CSS
- Markdown
- JSON/YAML
- And many more...

The editor automatically detects file types based on file extensions and applies appropriate syntax highlighting using the "doom-one" color scheme.

## Architecture

Gecko is built with a modular architecture:

- **`main.go`**: Core application logic and UI rendering
- **`textbuffer.go`**: Text manipulation and cursor management
- **`keybinds.go`**: Keyboard shortcut definitions
- **`minibuffer.go`**: Interactive dialogs and search interface
- **`modelupdate.go`**: Event handling and state updates
- **`syntax.go`**: Syntax highlighting integration

## Roadmap

- [ ] Configuration file support
- [ ] Theme customization
- [ ] Multiple file tabs
- [ ] Split pane editing
- [ ] Plugin system
- [ ] Advanced search and replace
- [ ] Git integration
- [ ] Line wrapping options
- [ ] Bracket matching
- [ ] Auto-completion


## Known Issues

- Clipboard integration requires `xclip` on Linux systems
- Some terminal emulators may not support all key combinations
- Very large files (>10MB) may experience performance issues

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - The amazing TUI framework
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) - Terminal styling
- [Chroma](https://github.com/alecthomas/chroma) - Syntax highlighting
- [Bubbles](https://github.com/charmbracelet/bubbles) - TUI components

---

<div align="center">
  <strong>🦎 Built with Go and ❤️</strong>
</div>
