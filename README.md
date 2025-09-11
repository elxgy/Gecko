# Gecko - Terminal Text Editor

Gecko is a lightweight, terminal-based text editor built with Go and the Bubble Tea framework. It provides essential text editing capabilities with syntax highlighting for various programming languages, designed for developers who prefer working in the terminal.

## Features

### Core Editing
- **Text Editing**: Basic text input, deletion, and modification with efficient text buffer management
- **Cursor Movement**: Navigate through text using arrow keys, Home, End, Page Up/Down
- **Line Operations**: Insert new lines, join lines, and advanced line manipulation

### Text Operations
- **Selection**: Select text using Shift + arrow keys with visual feedback
- **Copy/Cut/Paste**: Standard clipboard operations (Ctrl+C, Ctrl+X, Ctrl+V)
- **Undo/Redo**: Multi-level undo and redo functionality (Ctrl+Z, Ctrl+Y)
- **Search & Replace**: Find and replace text with regex support

### UI Features
- **Syntax Highlighting**: Support for 200+ programming languages via Chroma
- **Line Numbers**: Display line numbers in the left margin
- **Status Bar**: Show current cursor position, file status, mode, and file type
- **Responsive Interface**: Adapts to terminal size changes automatically
- **Theme Support**: Multiple color schemes for different preferences

### File Management
- **File Operations**: Open, save, and create new files with proper error handling
- **Multiple File Support**: Work with multiple files simultaneously
- **Auto-save**: Configurable automatic saving functionality
- **File Type Detection**: Automatic syntax highlighting based on file extension

## Prerequisites

- **Go 1.24.3** or later
- A terminal that supports ANSI escape sequences and 256 colors
- **Platform-specific dependencies**:
  - **Linux**: `xclip` or `wl-clipboard` for clipboard functionality
  - **macOS**: Built-in clipboard support
  - **Windows**: Built-in clipboard support

## Installation

### Windows

#### Option 1: Using Git Bash or WSL
```bash
# Clone the repository
git clone https://github.com/yourusername/gecko.git
cd gecko

# Build the application
go build -o gecko.exe

# Add to PATH (optional)
mkdir -p ~/bin
cp gecko.exe ~/bin/
echo 'export PATH="$HOME/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc
```

#### Option 2: Using PowerShell
```powershell
# Clone the repository
git clone https://github.com/yourusername/gecko.git
cd gecko

# Build the application
go build -o gecko.exe

# Add to PATH (optional)
$env:PATH += ";$(Get-Location)"
```

### macOS

#### Option 1: Using Homebrew (Recommended)
```bash
# Install Go if not already installed
brew install go

# Clone and build
git clone https://github.com/yourusername/gecko.git
cd gecko
go build -o gecko

# Install globally
sudo cp gecko /usr/local/bin/
```

#### Option 2: Manual Installation
```bash
# Download and install Go from https://golang.org/dl/
# Then clone and build
git clone https://github.com/yourusername/gecko.git
cd gecko
go build -o gecko

# Add to PATH
echo 'export PATH="$PWD:$PATH"' >> ~/.zshrc
source ~/.zshrc
```

### Linux

#### Ubuntu/Debian
```bash
# Install dependencies
sudo apt update
sudo apt install golang-go git xclip

# Clone and build
git clone https://github.com/yourusername/gecko.git
cd gecko
go build -o gecko

# Install globally
sudo cp gecko /usr/local/bin/
```

#### Arch Linux
```bash
# Install dependencies
sudo pacman -S go git xclip

# Clone and build
git clone https://github.com/yourusername/gecko.git
cd gecko
go build -o gecko

# Install globally
sudo cp gecko /usr/local/bin/
```

#### Fedora/RHEL/CentOS
```bash
# Install dependencies
sudo dnf install golang git xclip  # Fedora
# OR
sudo yum install golang git xclip  # RHEL/CentOS

# Clone and build
git clone https://github.com/yourusername/gecko.git
cd gecko
go build -o gecko

# Install globally
sudo cp gecko /usr/local/bin/
```

#### Generic Linux
```bash
# Download Go from https://golang.org/dl/
wget https://go.dev/dl/go1.24.3.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.24.3.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin

# Install clipboard utility
# For X11: sudo apt/yum/pacman install xclip
# For Wayland: sudo apt/yum/pacman install wl-clipboard

# Clone and build
git clone https://github.com/yourusername/gecko.git
cd gecko
go build -o gecko
sudo cp gecko /usr/local/bin/
```

## How to Use the Editor

### Getting Started

#### Opening Files
```bash
# Create a new file
gecko

# Open an existing file
gecko filename.txt

# Open multiple files
gecko file1.txt file2.txt file3.py

# Open with specific syntax highlighting
gecko --syntax=python script.py
```

#### First Steps Tutorial
1. **Start Gecko**: Run `gecko` or `gecko filename.txt`
2. **Navigate**: Use arrow keys to move the cursor
3. **Edit Text**: Simply start typing to insert text
4. **Save**: Press `Ctrl+S` to save your work
5. **Exit**: Press `Ctrl+Q` to quit (you'll be prompted to save unsaved changes)

### Basic Workflow

#### Text Editing
- **Insert Mode**: Default mode - type to insert text
- **Selection**: Hold `Shift` + arrow keys to select text
- **Copy/Cut/Paste**: Use `Ctrl+C`, `Ctrl+X`, `Ctrl+V`
- **Undo/Redo**: Use `Ctrl+Z` to undo, `Ctrl+Y` to redo

#### File Operations
- **New File**: `Ctrl+N`
- **Open File**: `Ctrl+O`
- **Save**: `Ctrl+S`
- **Save As**: `Ctrl+Shift+S`
- **Close File**: `Ctrl+W`

#### Navigation
- **Line Start/End**: `Home`/`End`
- **File Start/End**: `Ctrl+Home`/`Ctrl+End`
- **Page Up/Down**: `Page Up`/`Page Down`
- **Go to Line**: `Ctrl+G`

### Advanced Features

#### Search and Replace
- **Find**: `Ctrl+F` - Search for text
- **Find Next**: `F3` or `Ctrl+G`
- **Replace**: `Ctrl+H` - Find and replace
- **Regex Search**: Enable regex mode in search dialog

#### Multiple Files
- **Switch Files**: `Ctrl+Tab` to cycle through open files
- **File List**: `Ctrl+Shift+O` to see all open files
- **Close Current**: `Ctrl+W`
- **Close All**: `Ctrl+Shift+W`

### Tips and Tricks

1. **Syntax Highlighting**: Gecko automatically detects file types based on extensions
2. **Auto-save**: Enable auto-save in settings to prevent data loss
3. **Terminal Compatibility**: Works best with terminals supporting 256 colors
4. **Large Files**: Gecko efficiently handles files up to several MB
5. **Clipboard**: Ensure clipboard utilities are installed for copy/paste functionality

### Troubleshooting

#### Common Issues

**Clipboard not working on Linux:**
```bash
# Install xclip for X11
sudo apt install xclip

# Or wl-clipboard for Wayland
sudo apt install wl-clipboard
```

**Colors not displaying correctly:**
- Ensure your terminal supports 256 colors
- Try setting `TERM=xterm-256color`

**Performance issues with large files:**
- Gecko is optimized for files up to 10MB
- For larger files, consider using streaming editors

**Build errors:**
```bash
# Ensure Go version is correct
go version  # Should be 1.24.3 or later

# Clean and rebuild
go clean
go mod tidy
go build
```

### Keyboard Shortcuts

| Action | Shortcut |
|--------|----------|
| **File Operations** |
| New File | `Ctrl+N` |
| Open File | `Ctrl+O` |
| Save | `Ctrl+S` |
| Save As | `Ctrl+Shift+S` |
| Close File | `Ctrl+W` |
| Close All | `Ctrl+Shift+W` |
| Quit | `Ctrl+Q` |
| **Navigation** |
| Move cursor | `Arrow Keys` |
| Jump to line start | `Home` |
| Jump to line end | `End` |
| Jump to file start | `Ctrl+Home` |
| Jump to file end | `Ctrl+End` |
| Page up/down | `Page Up/Down` |
| Go to line | `Ctrl+G` |
| Switch files | `Ctrl+Tab` |
| File list | `Ctrl+Shift+O` |
| **Text Operations** |
| Select text | `Shift + Arrow Keys` |
| Select word | `Ctrl+Shift + Arrow Keys` |
| Select all | `Ctrl+A` |
| Copy | `Ctrl+C` |
| Cut | `Ctrl+X` |
| Paste | `Ctrl+V` |
| Undo | `Ctrl+Z` |
| Redo | `Ctrl+Y` |
| **Search & Replace** |
| Find | `Ctrl+F` |
| Find next | `F3` or `Ctrl+G` |
| Replace | `Ctrl+H` |
| **Other** |
| Show help | `F1` or `Ctrl+?` |

## Syntax Highlighting

Gecko supports syntax highlighting for 200+ programming languages through the Chroma library:

- **Web Technologies**: HTML, CSS, JavaScript, TypeScript, PHP, Vue, React (JSX)
- **Systems Programming**: C, C++, Rust, Go, Zig, Assembly
- **Scripting Languages**: Python, Ruby, Perl, Bash, PowerShell, Lua
- **JVM Languages**: Java, Kotlin, Scala, Clojure, Groovy
- **Functional Languages**: Haskell, F#, Erlang, Elixir, OCaml
- **Data Formats**: JSON, YAML, TOML, XML, CSV, SQL
- **Markup Languages**: Markdown, LaTeX, reStructuredText, AsciiDoc
- **Configuration**: Dockerfile, Nginx, Apache, INI
- **And many more...**

Syntax highlighting is automatically applied based on file extension. You can also manually specify the language using the `--syntax` flag.

## Project Structure

Gecko follows a modular architecture built on the Bubble Tea framework:

```
gecko/
├── .github/
│   └── workflows/
│       ├── ci.yml                    # Continuous Integration workflow
│       ├── commit-lint.yml           # Commit message linting
│       └── branch-protection.yml     # Branch protection automation
├── main.go                           # Application entry point and CLI handling
├── model.go                          # Main Bubble Tea model and state management
├── textbuffer.go                     # Core text buffer implementation
├── ui.go                             # User interface rendering and styling
├── syntax.go                         # Syntax highlighting integration
├── clipboard.go                      # Clipboard operations (cross-platform)
├── selection.go                      # Text selection handling
├── .golangci.yml                     # Go linting configuration
├── commitlint.config.js              # Commit message format rules
├── go.mod                            # Go module dependencies
├── go.sum                            # Dependency checksums
├── README.md                         # Project documentation
└── LICENSE                           # MIT License
```

### Core Components

- **Text Buffer (`textbuffer.go`)**: Efficient rope-based data structure for handling large files with O(log n) operations
- **UI Layer (`ui.go`, `model.go`)**: Bubble Tea components for rendering and user interaction
- **Syntax Engine (`syntax.go`)**: Chroma integration for language-specific highlighting with theme support
- **Selection System (`selection.go`)**: Advanced text selection with visual feedback and multi-line support
- **Clipboard Manager (`clipboard.go`)**: Cross-platform clipboard operations with fallback mechanisms
- **CI/CD Pipeline**: Automated testing, linting, security scanning, and quality assurance

## Future Features

Gecko has an exciting roadmap of features planned for future releases:

### File Navigation & Management
- **File Tree Navigation**: Built-in file explorer with tree view
- **Fuzzy File Finder**: Quick file opening with fuzzy search
- **Project Management**: Workspace support with project-specific settings
- **File Tabs**: Visual tabs for managing multiple open files
- **Split Panes**: Horizontal and vertical split editing

### LSP Integration & Smart Features
- **Language Server Protocol (LSP)**: Full LSP client implementation
- **Autocompletion**: Intelligent code completion based on context
- **Go to Definition**: Navigate to symbol definitions
- **Hover Information**: Display documentation and type information
- **Diagnostics**: Real-time error and warning display
- **Code Actions**: Quick fixes and refactoring suggestions
- **Symbol Search**: Find symbols across the project

### Advanced Editing Features
- **Multiple Cursors**: Edit multiple locations simultaneously
- **Code Folding**: Collapse and expand code blocks
- **Bracket Matching**: Highlight matching brackets and parentheses
- **Auto-indentation**: Smart indentation based on language rules
- **Snippet Support**: Expandable code templates
- **Macro Recording**: Record and replay editing sequences

### Developer Tools Integration
- **Git Integration**: Built-in git status, diff view, and blame
- **Debugger Support**: Integration with language-specific debuggers
- **Terminal Integration**: Embedded terminal for running commands
- **Build System**: Integration with common build tools
- **Testing Framework**: Run and display test results

### Customization & Extensibility
- **Plugin System**: Extensible architecture for custom plugins
- **Theme Engine**: Customizable color schemes and UI themes
- **Keybinding Customization**: User-defined keyboard shortcuts
- **Configuration Management**: Comprehensive settings system
- **Workspace Settings**: Project-specific configurations

### Performance & Scalability
- **Large File Handling**: Optimized for files up to 1GB
- **Streaming Mode**: Handle extremely large files with streaming
- **Background Processing**: Non-blocking operations for better responsiveness
- **Memory Optimization**: Efficient memory usage for large codebases

## Roadmap Timeline

### Q1 2025: Foundation & Navigation
- [ ] File tree navigation implementation
- [ ] Basic LSP client architecture
- [ ] Configuration file support
- [ ] Enhanced search with regex

### Q2 2025: LSP Integration
- [ ] LSP autocompletion
- [ ] Go to definition and hover
- [ ] Diagnostics display
- [ ] Symbol search

### Q3 2025: Advanced Features
- [ ] Multiple cursors
- [ ] Code folding
- [ ] Git integration
- [ ] Plugin system foundation

### Q4 2025: Polish & Performance
- [ ] Theme engine
- [ ] Performance optimizations
- [ ] Large file handling
- [ ] Comprehensive testing


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
