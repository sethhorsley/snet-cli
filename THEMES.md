# Themes

The snet TUI supports multiple color themes with live switching.

## Default Theme

**Default (Transparent)** - A clean purple/cyan theme with transparent background that works great with any terminal theme.

## Changing Themes

While the TUI is running, press `T` (capital T) to open the theme selector.

The theme selector appears as an **overlay** on top of the main interface, showing you the UI behind it with the **current theme applied in real-time** as you navigate.

### Theme Selector Controls

- `↑`/`↓`, `j`/`k`, or `Ctrl+P`/`Ctrl+N` - Navigate and **preview themes live** (theme changes as you move)
- `Enter` or `T` - **Apply** the currently previewed theme and close selector
- `Esc` - **Cancel** and revert to your previous theme

## Available Themes

### Default
- **Default (Transparent)** - Clean purple/cyan theme with transparent background (default)

### Catppuccin Family

- **Catppuccin Mocha** - Dark pastel theme
- **Catppuccin Latte** - Light pastel theme
- **Catppuccin Frappé** - Medium dark pastel theme
- **Catppuccin Macchiato** - Dark pastel theme

### Other Themes

- **Nord** - Arctic, north-bluish color palette
- **Dracula** - Dark theme with vibrant colors
- **Gruvbox Dark** - Retro groove dark theme
- **Tokyo Night** - A clean, dark theme inspired by Tokyo at night
- **Solarized Dark** - Precision colors for professionals

## Preview

When you open the theme selector with `T`, you'll see:

```
Select Theme

▸ Default (Transparent) ✓
  Catppuccin Mocha
  Catppuccin Latte
  Catppuccin Frappé
  Catppuccin Macchiato
  Nord
  Dracula
  Gruvbox Dark
  Tokyo Night
  Solarized Dark

↑/↓ j/k Ctrl+P/N: preview  •  Enter/T: apply  •  Esc: cancel
```

The `▸` shows your current selection and `✓` shows the currently saved theme.

### How Live Preview Works

1. Press `T` to open the selector (starts at your current theme)
2. Press `↓` → **Entire UI instantly changes to the next theme**
3. Press `↓` again → **UI changes to the next theme** (and so on)
4. Like a theme? Press `Enter` or `T` to keep it
5. Don't like any? Press `Esc` to go back to your original theme

**The entire interface updates in real-time** - headers, colors, borders, request logs, status indicators - everything changes as you navigate!

## Theme Features

- **Live Preview** - Themes apply **instantly as you navigate** through the selector - see changes in real-time!
- **Overlay Selector** - Theme selector appears as an overlay on top of the main UI
- **Non-Destructive Preview** - Press `Esc` to cancel and revert to your previous theme
- **See Before You Apply** - The entire UI behind the selector updates live as you browse themes
- **Persistent Within Session** - Your theme choice stays active for the current tunnel session
- **Background Color Support** - Themes can have solid backgrounds or transparent backgrounds
- **Transparent Support** - The Default theme uses a transparent background
- **Emacs/Vim Navigation** - Supports `Ctrl+P`/`Ctrl+N` (Emacs) and `j`/`k` (Vim) in addition to arrow keys
- **Semantic Colors** - All themes maintain consistent meaning for colors:
  - Success/2xx responses: Green
  - Info/3xx responses: Blue/Cyan
  - Warning/4xx responses: Yellow
  - Error/5xx responses: Red
  - Primary UI elements: Theme's primary color
  - Muted text: Gray tones

## Examples

### Default (Transparent)
- Clean purple and cyan accents
- High contrast for readability
- Transparent background blends with your terminal

### Tokyo Night
- Cool blue tones
- Perfect for late-night coding
- Vibrant but easy on the eyes

### Dracula
- High contrast
- Vibrant purples and cyans
- Great for presentation mode

## Adding Custom Themes

Themes are defined in `internal/tui/themes.go`. To add your own theme, add a new entry to the `themes` map and include it in `ThemeNames()`.

Each theme defines:
- Primary (main accent color)
- Secondary (secondary accent)
- Success (green tones)
- Warning (yellow tones)
- Error (red tones)
- Muted (gray tones for less important text)
- Border (border colors)
- Text (main text color)
- Background (background color, can be empty for transparent)
- UseTransparentBg (whether to use transparent background)
