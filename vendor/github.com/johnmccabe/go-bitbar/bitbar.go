// Copyright (c) John McCabe 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

// Package bitbar simplifies the creation of Bitbar plugins.
//
// Provides helper functions for adding lines and submenus along with setting
// command, style options etc using function chaining.
//
// See the BitBar project for more info - https://github.com/matryer/bitbar
package bitbar

import (
	"fmt"
	"image"
	"strings"
)

// Plugin holds the content of the Bitbar plugin, lines and submenus.
type Plugin struct {
	StatusBar StatusBar
	SubMenu   *SubMenu
}

// Line holds the content, styling and behaviour of a line in a Bitbar
// menu, both in the menu and submenus
type Line struct {
	text          string
	href          string
	color         string
	font          string
	size          int
	terminal      *bool
	refresh       *bool
	dropDown      *bool
	length        int
	trim          *bool
	alternate     *bool
	emojize       *bool
	ansi          *bool
	bash          string
	params        []string
	templateImage string
	image         string
	hr            bool
}

// Style wraps options related to text presentation which can be added to a line
// using the *line.Style(s Style) function.
type Style struct {
	Color   string
	Font    string
	Size    int
	Length  int
	Trim    *bool
	Emojize *bool
	Ansi    *bool
}

// Cmd wraps options related to commands which can be added to a line using the
// *line.Command(c Cmd) function.
type Cmd struct {
	Bash     string
	Params   []string
	Terminal *bool
	Refresh  *bool
}

// StatusBar holds one of more Lines of text which are rendered in the status bar.
// Multiple Lines will be cycled through over and over
type StatusBar struct {
	Lines []*Line
}

// SubMenuItem is used to hold a Line or SubMenu.
type SubMenuItem interface{}

// SubMenu contains a slice of SubMenuItems which can be Lines or additional
// SubMenus. The Level indicates how nested the submenu is which is used during
// render to prepend the correct number of `--` prefixes.
type SubMenu struct {
	Level int
	Lines []SubMenuItem
}

// New returns an empty Bitbar menu without any context
func New() Plugin {
	return Plugin{}
}

// StatusLine creates a line adding text to the status bar which will be added
// before the main dropdown delimiter (`---`), multiple StatusLines will be
// cycled through over and over.
//  *menu.StatusLine("Text for the status bar")
func (p *Plugin) StatusLine(s string) *Line {
	l := new(Line)
	l.text = s
	p.StatusBar.Lines = append(p.StatusBar.Lines, l)
	return l
}

// NewSubMenu creates a submenu off the main menu.
//  *menu.NewSubMenu()
func (p *Plugin) NewSubMenu() *SubMenu {
	p.SubMenu = new(SubMenu)
	p.SubMenu.Level = 0
	return p.SubMenu
}

// Line creates a line adding text to the dropdown which will be added after
// the main dropdown delimiter (`---`).
//  submenu.Line("Submenu item text")
func (d *SubMenu) Line(s string) *Line {
	l := new(Line)
	l.text = s
	d.Lines = append(d.Lines, l)
	return l
}

// Image adds a line with an image to the dropdown which will be added after
// the main dropdown delimiter (`---`). Use a 144 DPI resolution to support
// Retina displays.
//  line.Image(myImg)
func (d *SubMenu) Image(img image.Image) *Line {
	return d.Line("").Image(img)
}

// HR turns a line into a horizontal delimiter, useful for breaking menu items
// into logical groups.
//  submenu.Line("").HR()
func (d *SubMenu) HR() *Line {
	l := new(Line)
	l.hr = true
	l.text = "---"
	d.Lines = append(d.Lines, l)
	return l
}

// NewSubMenu creates a nested submenu off a submenu.
//  submenu.NewSubMenu()
func (d *SubMenu) NewSubMenu() *SubMenu {
	newSubMenu := new(SubMenu)
	newSubMenu.Level = d.Level + 1
	d.Lines = append(d.Lines, newSubMenu)
	return newSubMenu
}

// Style provides a alternate method for setting the text style related
// options.
//  style := bitbar.Style{
//    Color:   "red",
//    Font:    "UbuntuMono-Bold",
//    Size:    14,
//    Length:  20,
//    Trim:    false,
//    Emojize: false,
//    Ansi:    false,
//  }
//  line.Style(false)
func (l *Line) Style(s Style) *Line {
	l.color = s.Color
	l.font = s.Font
	l.size = s.Size
	l.length = s.Length
	l.trim = s.Trim
	l.emojize = s.Emojize
	l.ansi = s.Ansi
	return l
}

// Command provides a alternate method for setting the bash script and
// params along with some related flags via a Command struct.
//  cmd := bitbar.Cmd{
//    Bash:     "/Users/user/BitBar_Plugins/scripts/nginx.restart.sh",
//    Params:   []string{"--verbose"},
//    Terminal: false,
//    Refresh:  true,
//  }
//  line.Command(cmd)
func (l *Line) Command(c Cmd) *Line {
	l.bash = c.Bash
	l.params = c.Params
	l.terminal = c.Terminal
	l.refresh = c.Refresh
	return l
}

// Href adds a URL to the line and makes it clickable.
//  line.Href("http://github.com/johnmccabe/bitbar")
func (l *Line) Href(s string) *Line {
	l.href = s
	return l
}

// Color sets the lines font color, can take a name or hex value.
//  line.Color("red")
//  line.Color("#ff0000")
func (l *Line) Color(s string) *Line {
	l.color = s
	return l
}

// Font sets the lines font.
//  line.Font("UbuntuMono-Bold")
func (l *Line) Font(s string) *Line {
	l.font = s
	return l
}

// Size sets the lines font size.
//  line.Size(12)
func (l *Line) Size(i int) *Line {
	l.size = i
	return l
}

// Bash makes makes the line clickable and adds a script that will be run
// on click.
//  line.Bash("/Users/user/BitBar_Plugins/scripts/nginx.restart.sh")
func (l *Line) Bash(s string) *Line {
	l.bash = s
	return l
}

// Params adds arguments which are passed to the script specified by *Line.Bash().
//  args := []string{"--verbose"}
//  line.Bash("/Users/user/BitBar_Plugins/scripts/nginx.restart.sh").Params(args)
func (l *Line) Params(s []string) *Line {
	l.params = s
	return l
}

// Terminal sets a flag which controls whether a Terminal is opened when the bash
// script is run.
//  line.Bash("/Users/user/BitBar_Plugins/scripts/nginx.restart.sh").Terminal(true)
func (l *Line) Terminal(b bool) *Line {
	l.terminal = &b
	return l
}

// Refresh controls whether clicking the line results in the plugin being refreshed.
// If the line has a bash script attached then the plugin is refreshed after the
// script finishes.
//  line.Bash("/Users/user/BitBar_Plugins/scripts/nginx.restart.sh").Refresh()
//  line.Refresh()
func (l *Line) Refresh() *Line {
	refreshEnabled := true
	l.refresh = &refreshEnabled
	return l
}

// DropDown sets a flag which controls whether the line only appears and cycles in the
// status bar but not in the dropdown.
//  line.DropDown(false)
func (l *Line) DropDown(b bool) *Line {
	l.dropDown = &b
	return l
}

// Length truncates the line after the specified number of characters. An elipsis will
// be added to any truncated strings, as well as a tooltip displaying the full string.
//  line.DropDown(false)
func (l *Line) Length(i int) *Line {
	l.length = i
	return l
}

// Trim sets a flag to control whether leading/trailing whitespace is trimmed from the
// title. Defaults to `true`.
//  line.Trim(false)
func (l *Line) Trim(b bool) *Line {
	l.trim = &b
	return l
}

// Alternate sets a flag to mark a line as an alternate to the previous one for when the
// Option key is pressed in the dropdown.
//  line.Alternate(false)
func (l *Line) Alternate(b bool) *Line {
	l.alternate = &b
	return l
}

// TemplateImage sets an image for the line. The image data must be passed as base64
// encoded string and should consist of only black and clear pixels. The alpha channel
// in the image can be used to adjust the opacity of black content, however. This is the
// recommended way to set an image for the statusbar. Use a 144 DPI resolution to support
// Retina displays. The imageformat can be any of the formats supported by Mac OS X.
//  line.TemplateImage("iVBORw0KGgoAAAANSUhEUgAAA...")
func (l *Line) TemplateImage(s string) *Line {
	l.templateImage = s
	return l
}

// Image set an image for the line. Use a 144 DPI resolution to support Retina displays.
//  line.Image(myImg)
func (l *Line) Image(img image.Image) *Line {
	l.image = toBase64(img)
	return l
}

// Emojize sets a flag to control parsing of github style :mushroom: into 🍄.
//  line.Emojize(false)
func (l *Line) Emojize(b bool) *Line {
	l.emojize = &b
	return l
}

// Ansi sets a flag to control parsing of ANSI codes.
//  line.Ansi(false)
func (l *Line) Ansi(b bool) *Line {
	l.ansi = &b
	return l
}

// CopyToClipboard is a helper to copy the specified text to the OSX clipboard.
//  line.CopyToClipboard("some text")
func (l *Line) CopyToClipboard(text string) *Line {
	line := l.Bash("/bin/bash").
		Params([]string{"-c", fmt.Sprintf("'echo -n %s | pbcopy'", text)}).
		Terminal(false)
	return line
}

// Render the Bitbar menu to Stdout.
func (p *Plugin) Render() {
	var output string
	for _, line := range p.StatusBar.Lines {
		output = output + fmt.Sprintf("%s\n", renderLine(line))
	}
	output = output + "---\n"
	if p.SubMenu != nil {
		output = output + renderSubMenu(p.SubMenu)
	}
	fmt.Println(output)
}

func renderSubMenu(d *SubMenu) string {
	var output string
	var prefix string
	if d.Level > 0 {
		prefix = strings.Repeat("--", d.Level) + " "
	}
	for _, line := range d.Lines {
		switch v := line.(type) {
		case *Line:
			if line.(*Line).hr {
				output = output + fmt.Sprintf("%s%s\n", strings.TrimSpace(prefix), renderLine(v))
			} else {
				output = output + fmt.Sprintf("%s%s\n", prefix, renderLine(v))
			}
		case *SubMenu:
			output = output + renderSubMenu(v)
		}
	}
	return output
}

func renderLine(l *Line) string {
	result := []string{l.text}
	options := []string{"|"}
	options = append(options, renderStyleOptions(l)...)
	options = append(options, renderCommandOptions(l)...)
	options = append(options, renderImageOptions(l)...)
	options = append(options, renderMiscOptions(l)...)

	if len(options) > 1 {
		result = append(result, options...)
	}

	return strings.Join(result, " ")
}

func renderImageOptions(l *Line) []string {
	imageOptions := []string{}
	if len(l.templateImage) > 0 {
		imageOptions = append(imageOptions, fmt.Sprintf("templateImage='%s'", l.templateImage))
	}
	if len(l.image) > 0 {
		imageOptions = append(imageOptions, fmt.Sprintf("image='%s'", l.image))
	}

	return imageOptions
}

func renderMiscOptions(l *Line) []string {
	miscOptions := []string{}
	if l.href != "" {
		miscOptions = append(miscOptions, fmt.Sprintf("href='%s'", l.href))
	}
	if l.dropDown != nil {
		miscOptions = append(miscOptions, fmt.Sprintf("dropdown=%t", *l.dropDown))
	}
	if l.alternate != nil {
		miscOptions = append(miscOptions, fmt.Sprintf("alternate='%t'", *l.alternate))
	}

	return miscOptions
}

func renderStyleOptions(l *Line) []string {
	styleOptions := []string{}
	if l.color != "" {
		styleOptions = append(styleOptions, fmt.Sprintf("color=\"%s\"", l.color))
	}
	if l.font != "" {
		styleOptions = append(styleOptions, fmt.Sprintf("font=\"%s\"", l.font))
	}
	if l.size > 0 {
		styleOptions = append(styleOptions, fmt.Sprintf("size=%d", l.size))
	}
	if l.length > 0 {
		styleOptions = append(styleOptions, fmt.Sprintf("length=%d", l.length))
	}
	if l.trim != nil {
		styleOptions = append(styleOptions, fmt.Sprintf("trim=%t", *l.trim))
	}
	if l.emojize != nil {
		styleOptions = append(styleOptions, fmt.Sprintf("emojize=%t", *l.emojize))
	}
	if l.ansi != nil {
		styleOptions = append(styleOptions, fmt.Sprintf("ansi=%t", *l.ansi))
	}
	return styleOptions
}

func renderCommandOptions(l *Line) []string {
	commandOptions := []string{}
	if l.bash != "" {
		commandOptions = append(commandOptions, fmt.Sprintf("bash=\"%s\"", l.bash))
	}
	if len(l.params) > 0 {
		for i, param := range l.params {
			commandOptions = append(commandOptions, fmt.Sprintf("param%d=%s", i+1, param))
		}
	}
	if l.terminal != nil {
		commandOptions = append(commandOptions, fmt.Sprintf("terminal=%t", *l.terminal))
	}
	if l.refresh != nil {
		commandOptions = append(commandOptions, fmt.Sprintf("refresh=%t", *l.refresh))
	}

	return commandOptions
}
