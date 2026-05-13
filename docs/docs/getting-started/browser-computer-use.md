# Browser Use and Computer Use

Memoh can give a bot a visible workspace desktop and a headed browser inside its workspace container. This is different from running a headless Playwright script: the bot can inspect and operate the same graphical browser that you can see in the Web UI display pane.

## Concepts

| Capability | Best for | How it works |
|------------|----------|--------------|
| Headless browser commands | Fast scripted automation inside a workspace | Run Playwright or other browser tooling as normal workspace commands. |
| Browser Use | Web pages, forms, navigation, screenshots, accessibility-tree inspection | Operates the headed workspace Chrome/Chromium instance over CDP. |
| Computer Use | Native dialogs, broken browser state, non-browser GUI, coordinate-level recovery | Uses desktop screenshots plus pointer and keyboard input. |

Prefer Browser Use for web pages. Use Computer Use when the task depends on GUI state that CDP cannot reach.

## Workspace display and VNC

Workspace display is the desktop environment inside the bot's workspace container. VNC/RFB is the display and input transport behind that desktop, while WebRTC is used by the Web UI display session.

The main value is not VNC by itself. The important capability is that the workspace can run a headed Chrome/Chromium browser for sites and login flows that do not work well in headless mode.

## Preparing a bot desktop

1. Open the bot detail page.
2. Go to the **Desktop** tab.
3. Prepare or enable the workspace display runtime.
4. Open a display session from the bot settings page or from the chat workspace.

The display runtime installs or uses the workspace desktop, VNC server, browser, and fonts needed for the visible session. Availability depends on the workspace backend and image.

## Agent tools

When workspace desktop is enabled, the agent can use browser and computer tools:

- `browser_observe` inspects the current browser page.
- `browser_action` clicks, fills, types, presses, and navigates in the headed browser.
- `browser_remote_session` exposes the browser CDP endpoint for code-driven sessions.
- `computer_use` takes screenshots and sends pointer or keyboard input to the desktop.

These tools are workspace runtime features. They do not automate the Electron desktop app itself.

## Related

- [Bot Workspace Management](/getting-started/container)
- [Workspace Backends](/installation/workspace-backends)
