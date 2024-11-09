# Verilog Language Server Extension

Inspired by the lack of editor support in existing verilog extensions for VSCode, this repository contains a powerful, hand-written Verilog language server and an extension to use it with VSCode.

## Functionality

This Language Server works for Verilog files.

Current Features:

- Completions
- Go To Definition
- Token Highlighting

Roadmap Features:

- Parsing Module Insides
- Error-tolerant parser
- Improved Completions
- Diagnostics

## Run in Development Environment

- Clone this repository
- Run `npm install` in the repo. This installs all necessary npm modules in both the client and server folder.
- Open the repo in VS Code
- Switch to the Run and Debug View in the Sidebar (shortcut: <kbd>Ctrl+Shift+D</kbd>)
- Select `VLS` from the drop down
  - This will launch both the [Extension Development Host](https://code.visualstudio.com/api/get-started/your-first-extension#:~:text=Then%2C%20inside%20the%20editor%2C%20press%20F5.%20This%20will%20compile%20and%20run%20the%20extension%20in%20a%20new%20Extension%20Development%20Host%20window.) in debug mode, and the language server in Golang Debugger
  - Set breakpoints any source file in `server/` or `client/`
- Press ▷ to run the launch config (shortcut: <kbd>F5</kbd>)
- In the Extension Development Host, open a Verilog Document
  - You should now be able to use the extension and see the autocomplete functionality
