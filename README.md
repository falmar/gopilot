# WIP: gopilot

An attempt to run Chromium automation with bare CDP commands.

> **NOTE:** Breaking changes may occur until the API is finalized.

## Overview

gopilot is my attempt to provide a simple, minimalistic API for automating Chromium browsers. It's not meant to be another Puppeteer. Instead, it's focused on the essential features most users need for straightforward browser tasks—no fluff, just what you need.

## Why Minimalistic?

I wanted to simplify browser automation by sticking to the core functionalities that most of us use:
- Navigation to web pages
- Clicking on elements
- Typing text
- Taking screenshots
- Extracting HTML content

I’ve also added some features for intercepting requests, which is handy if you want to cancel or grab AJAX info. Overall, gopilot aims to be a lightweight tool that doesn’t bog you down with unnecessary complexity.

## Current Features

- **Navigate** to a specified URL
- **Click** on elements
- **Extract** HTML content from the page
- **Intercept** network requests for those who want to dig deeper
- **Set**, **get**, and **clear** cookies

### TODO:

- Taking screenshots of web pages
- Setting, getting, and clearing local storage
- Typing text into input fields

## Contributions

Contributions are welcome! If you've got a feature request or an idea to share, reach out. Remember to aimi for simplicity!
