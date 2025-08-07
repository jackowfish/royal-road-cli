# royal-road-cli

Terminal reader for Royal Road web novels.

## Install

```bash
go build -o royal-road-cli
./royal-road-cli
```

## Usage

```bash
# Interactive menu
royal-road-cli

# Browse popular fictions
royal-road-cli browse

# Read by fiction ID
royal-road-cli read [fiction-id]

# Continue where you left off
royal-road-cli continue
```

## Keys

### Reader
- `Space/f/j/l/→/↓` - Next page
- `k/h/←/↑` - Previous page  
- `n/b` - Next chapter
- `p` - Previous chapter
- `t` - Table of contents
- `m` - Main menu
- `r` - Reload chapter
- `?` - Help
- `q` - Quit

### Browse
- `Enter` - Select fiction
- `r` - Refresh list
- `q` - Quit

### Menu
- `c` - Continue reading
- `h` - History
- `n` - New book
- `b` - Browse
- `q` - Quit

## Requirements

- Go 1.21+

## Build

```bash
go build
```