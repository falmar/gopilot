// Package gopilot provides a simple and minimalistic API for automating Chromium browsers.
//
// gopilot is a lightweight alternative to complex browser automation tools, focusing on
// essential functionality using the Chrome DevTools Protocol (CDP). It's structured around
// three main components: Browser (manages instances), Page (represents tabs), and Element
// (represents DOM elements).
//
// Key features include navigation, DOM manipulation, element interaction, screenshots,
// and network request monitoring. The package supports both headful (default) and headless
// modes, and can be configured via environment variables like GOPILOT_CHROME_EXECUTABLE.
//
// Common use cases include web scraping, UI testing, form automation, and taking screenshots.
//
// For examples and detailed usage, see: https://github.com/falmar/gopilot/tree/main/examples
package gopilot
